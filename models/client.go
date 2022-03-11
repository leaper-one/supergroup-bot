package models

import (
	"context"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const client_DDL = `
-- 机器人信息
CREATE TABLE IF NOT EXISTS client (
	client_id          VARCHAR(36) NOT NULL PRIMARY KEY,
	identity_number    VARCHAR(11) NOT NULL DEFAULT '',
  client_secret      VARCHAR NOT NULL,
  session_id         VARCHAR(36) NOT NULL,
  pin_token          VARCHAR NOT NULL,
  private_key        VARCHAR NOT NULL,
  pin                VARCHAR(6) DEFAULT '',
  name               VARCHAR NOT NULL,
  icon_url           VARCHAR NOT NULL DEFAULT '',
  description        VARCHAR NOT NULL,
  host               VARCHAR NOT NULL, -- 前端部署的 host
	lang               VARCHAR NOT NULL DEFAULT 'zh',
  asset_id           VARCHAR(36) NOT NULL,
	owner_id					 VARCHAR(36) NOT NULL,
  speak_status       SMALLINT NOT NULL DEFAULT 1, -- 1 正常发言 2 持仓发言
	pay_status				 SMALLINT NOT NULL DEFAULT 0, -- 0 关闭 1 开启
	pay_amount			   VARCHAR NOT NULL DEFAULT '', -- 付费入群开启的金额
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type Client struct {
	ClientID       string    `json:"client_id,omitempty" redis:"client_id"`
	IdentityNumber string    `json:"identity_number,omitempty" redis:"identity_number"`
	ClientSecret   string    `json:"client_secret,omitempty" redis:"client_secret"`
	SessionID      string    `json:"session_id,omitempty" redis:"session_id"`
	PinToken       string    `json:"pin_token,omitempty" redis:"pin_token"`
	PrivateKey     string    `json:"private_key,omitempty" redis:"private_key"`
	Pin            string    `json:"pin,omitempty" redis:"pin"`
	Name           string    `json:"name,omitempty" redis:"name"`
	Description    string    `json:"description,omitempty" redis:"description"`
	Host           string    `json:"host,omitempty" redis:"host"`
	Lang           string    `json:"lang,omitempty" redis:"lang"`
	AssetID        string    `json:"asset_id,omitempty" redis:"asset_id"`
	OwnerID        string    `json:"owner_id,omitempty" redis:"owner_id"`
	SpeakStatus    int       `json:"speak_status,omitempty" redis:"speak_status"`
	PayStatus      int       `json:"pay_status,omitempty" redis:"pay_status"`
	PayAmount      string    `json:"pay_amount,omitempty" redis:"pay_amount"`
	CreatedAt      time.Time `json:"created_at,omitempty" redis:"created_at"`
	IconURL        string    `json:"icon_url,omitempty" redis:"icon_url"`
	Symbol         string    `json:"symbol,omitempty" redis:"symbol"`

	Welcome string `json:"welcome,omitempty" redis:"welcome"`
	JoinMsg string `json:"join_msg,omitempty" redis:"join_msg"`
}

const (
	ClientSpeckStatusOpen  = 1 // 持仓发言打开，
	ClientSpeckStatusClose = 2 // 持仓发言关闭

	ClientPayStatusOpen = 1 // 入群开启，
)

func UpdateClientSetting(ctx context.Context, u *ClientUser, desc, welcome string) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if desc != "" {
		if _, err := session.Database(ctx).Exec(ctx, `
UPDATE client SET description=$2 WHERE client_id=$1
`, u.ClientID, desc); err != nil {
			return err
		}
	}
	if welcome != "" {
		if err := updateClientWelcome(ctx, u.ClientID, welcome); err != nil {
			return err
		}
		go func() {
			// 给管理员发两条消息
			SendToClientManager(u.ClientID, &mixin.MessageView{
				ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
				UserID:         u.UserID,
				MessageID:      tools.GetUUID(),
				Category:       mixin.MessageCategoryPlainText,
				Data:           tools.Base64Encode([]byte(config.Text.WelcomeUpdate)),
				CreatedAt:      time.Now(),
			}, false, false)
			SendToClientManager(u.ClientID, &mixin.MessageView{
				ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
				UserID:         u.UserID,
				MessageID:      tools.GetUUID(),
				Category:       mixin.MessageCategoryPlainText,
				Data:           tools.Base64Encode([]byte(welcome)),
				CreatedAt:      time.Now(),
			}, false, false)
			go session.Redis(_ctx).Del(ctx, "client:"+u.ClientID)
		}()
	}
	return nil
}

func UpdateClient(ctx context.Context, c *Client) error {
	query := durable.InsertQueryOrUpdate("client", "client_id", "client_secret,session_id,pin_token,private_key,pin,name,icon_url,description,asset_id,host,speak_status,owner_id,identity_number,lang")
	go session.Redis(_ctx).Del(ctx, "client:"+c.ClientID)
	_, err := session.Database(ctx).Exec(ctx, query, c.ClientID, c.ClientSecret, c.SessionID, c.PinToken, c.PrivateKey, c.Pin, c.Name, c.IconURL, c.Description, c.AssetID, c.Host, c.SpeakStatus, c.OwnerID, c.IdentityNumber, c.Lang)
	return err
}

func GetClientByIDOrHost(ctx context.Context, clientIDorHost string) (Client, error) {
	var c Client
	key := "client:" + clientIDorHost
	if err := session.Redis(ctx).StructScan(ctx, key, &c); err != nil {
		if errors.Is(err, redis.Nil) {
			return cacheClient(ctx, clientIDorHost)
		}
		session.Logger(ctx).Println(err)
		return Client{}, err
	}
	return c, nil
}

func cacheClient(ctx context.Context, clientIDOrHost string) (Client, error) {
	var c Client
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT c.client_id,c.client_secret,c.session_id,c.pin_token,c.private_key,c.pin,c.host,c.asset_id,c.speak_status,c.created_at,c.name,c.description,c.icon_url,c.owner_id,c.pay_amount,c.pay_status,c.identity_number,c.lang,
cr.join_msg,cr.welcome
FROM client c
LEFT JOIN client_replay cr ON c.client_id=cr.client_id
WHERE c.client_id=$1 OR c.host=$1`,
		clientIDOrHost).Scan(&c.ClientID, &c.ClientSecret, &c.SessionID, &c.PinToken, &c.PrivateKey, &c.Pin, &c.Host, &c.AssetID, &c.SpeakStatus, &c.CreatedAt, &c.Name, &c.Description, &c.IconURL, &c.OwnerID, &c.PayAmount, &c.PayStatus, &c.IdentityNumber, &c.Lang,
		&c.JoinMsg, &c.Welcome); err != nil {
		return c, err
	}
	key1 := "client:" + c.ClientID
	key2 := "client:" + c.Host
	if err := session.Redis(ctx).StructSet(ctx, key1, c); err != nil {
		session.Logger(ctx).Println(err)
		return c, nil
	}
	if err := session.Redis(ctx).StructSet(ctx, key2, c); err != nil {
		session.Logger(ctx).Println(err)
		return c, nil
	}
	return c, nil
}

func GetClientList(ctx context.Context) ([]*Client, error) {
	clientList := make([]*Client, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id, session_id,pin_token,private_key,pin,speak_status,asset_id,created_at
FROM client
WHERE client_id=ANY($1)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var c Client
			if err := rows.Scan(&c.ClientID, &c.SessionID, &c.PinToken, &c.PrivateKey, &c.Pin, &c.SpeakStatus, &c.AssetID, &c.CreatedAt); err != nil {
				return err
			}
			clientList = append(clientList, &c)
		}
		return nil
	}, config.Config.ClientList)
	return clientList, err
}

func GetFirstClient(ctx context.Context) *mixin.Client {
	c, err := getAllClient(ctx)
	if err != nil {
		return nil
	}
	return GetMixinClientByIDOrHost(ctx, c[0]).Client
}

type MixinClient struct {
	*mixin.Client
	C Client
}

var cacheClientMap *tools.Mutex

func init() {
	cacheClientMap = tools.NewMutex()
}

func GetMixinClientByIDOrHost(ctx context.Context, clientIDOrHost string) MixinClient {
	client := cacheClientMap.Read(clientIDOrHost)
	if client == nil {
		c, err := GetClientByIDOrHost(ctx, clientIDOrHost)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				session.Logger(ctx).Println(err)
			}
			return *new(MixinClient)
		}
		client, err := mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   c.ClientID,
			SessionID:  c.SessionID,
			PinToken:   c.PinToken,
			PrivateKey: c.PrivateKey,
		})
		if err != nil {
			session.Logger(ctx).Println(err)
			return *new(MixinClient)
		}
		_client := MixinClient{
			Client: client,
			C:      c,
		}
		cacheClientMap.Write(clientIDOrHost, _client)
		return _client
	}
	return client.(MixinClient)
}

func GetClientStatusByID(ctx context.Context, u *ClientUser) string {
	return getClientConversationStatus(ctx, u.ClientID)
}

type ClientInfo struct {
	*Client
	PriceUsd      decimal.Decimal `json:"price_usd,omitempty" redis:"price_usd"`
	ChangeUsd     string          `json:"change_usd,omitempty" redis:"change_usd"`
	TotalPeople   decimal.Decimal `json:"total_people" redis:"total_people"`
	WeekPeople    decimal.Decimal `json:"week_people" redis:"week_people"`
	Activity      []*Activity     `json:"activity,omitempty"`
	HasReward     bool            `json:"has_reward" redis:"has_reward"`
	NeedPayAmount decimal.Decimal `json:"need_pay_amount,omitempty" redis:"need_pay_amount"`
	Amount        string          `json:"amount,omitempty" redis:"amount"`
	LargeAmount   string          `json:"large_amount,omitempty" redis:"large_amount"`
}

const (
	ExinOneClientID = "47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d"
	XinAssetID      = "c94ac88f-4671-3976-b60a-09064f1811e8"
	BtcAssetID      = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
)

func GetClientInfoByHostOrID(ctx context.Context, hostOrID string) (*ClientInfo, error) {
	mixinClient := GetMixinClientByIDOrHost(ctx, hostOrID)
	client := mixinClient.C
	var c ClientInfo
	if client.Pin != "" {
		c.HasReward = true
	}
	client.SessionID = ""
	client.PinToken = ""
	client.PrivateKey = ""
	client.Pin = ""
	c.Client = &client
	assetID := client.AssetID
	if c.ClientID == ExinOneClientID {
		assetID = XinAssetID
	} else if assetID == "" {
		assetID = BtcAssetID
	}
	asset, err := GetAssetByID(ctx, mixinClient.Client, assetID)
	if err == nil {
		c.PriceUsd = asset.PriceUsd
		c.ChangeUsd = asset.ChangeUsd
		c.Symbol = asset.Symbol
		if client.AssetID != "" && c.IconURL == "" {
			c.IconURL = asset.IconUrl
		}
	}
	amount, err := GetClientAssetLevel(ctx, client.ClientID)
	if err == nil {
		c.Amount = amount.Fresh.String()
		c.LargeAmount = amount.Large.String()
	}

	c.TotalPeople, c.WeekPeople, err = getClientPeopleCount(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}
	c.Activity, err = GetActivityByClientID(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetAllConfigClientInfo(ctx context.Context) ([]ClientInfo, error) {
	cis := make([]ClientInfo, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id FROM client WHERE client_id=ANY($1)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var clientID string
			if err := rows.Scan(&clientID); err != nil {
				return err
			}
			if ci, err := GetClientInfoByHostOrID(ctx, clientID); err != nil {
				return err
			} else {
				cis = append(cis, *ci)
			}
		}
		return nil
	}, config.Config.ShowClientList); err != nil {
		return nil, err
	}
	return cis, nil
}

func getAllClient(ctx context.Context) ([]string, error) {
	cs := make([]string, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id FROM client
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				return err
			}
			cs = append(cs, c)
		}
		return nil
	})
	return cs, err
}
