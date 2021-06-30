package models

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

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
  client_secret      VARCHAR NOT NULL,
  session_id         VARCHAR(36) NOT NULL,
  pin_token          VARCHAR NOT NULL,
  private_key        VARCHAR NOT NULL,
  pin                VARCHAR(6) DEFAULT '',
  name               VARCHAR NOT NULL,
  description        VARCHAR NOT NULL,
  host               VARCHAR NOT NULL, -- 前端部署的 host
  information_url    VARCHAR DEFAULT '',
  asset_id           VARCHAR(36) NOT NULL,
  speak_status       SMALLINT NOT NULL DEFAULT 1, -- 1 正常发言 2 持仓发言
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
)
`

type Client struct {
	ClientID       string    `json:"client_id,omitempty"`
	ClientSecret   string    `json:"client_secret,omitempty"`
	SessionID      string    `json:"session_id,omitempty"`
	PinToken       string    `json:"pin_token,omitempty"`
	PrivateKey     string    `json:"private_key,omitempty"`
	Pin            string    `json:"pin,omitempty"`
	Name           string    `json:"name,omitempty"`
	Description    string    `json:"description,omitempty"`
	Host           string    `json:"host,omitempty"`
	InformationURL string    `json:"information_url,omitempty"`
	AssetID        string    `json:"asset_id,omitempty"`
	SpeakStatus    int       `json:"speak_status,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`

	IconURL string `json:"icon_url,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
	Welcome string `json:"welcome,omitempty"`
	Amount  string `json:"amount,omitempty"`
}

const (
	ClientSpeckStatusOpen  = 1 // 持仓发言打开，
	ClientSpeckStatusClose = 2 // 持仓发言关闭
)

func UpdateClientSetting(ctx context.Context, u *ClientUser, desc, welcome string) error {
	if !checkIsManager(ctx, u.ClientID, u.UserID) {
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
	}
	cacheClient = make(map[string]Client)
	cacheClientReplay = make(map[string]ClientReplay)
	return nil
}

func UpdateClient(ctx context.Context, c *Client) error {
	if strings.HasSuffix(c.Host, "/") {
		c.Host = c.Host[:len(c.Host)-1]
	}
	query := durable.InsertQueryOrUpdate("client", "client_id", "client_secret,session_id,pin_token,private_key,pin,name,description,asset_id,host,information_url,speak_status")
	_, err := session.Database(ctx).Exec(ctx, query, c.ClientID, c.ClientSecret, c.SessionID, c.PinToken, c.PrivateKey, c.Pin, c.Name, c.Description, c.AssetID, c.Host, c.InformationURL, c.SpeakStatus)
	return err
}

var cacheClient = make(map[string]Client)
var nilClient = Client{}

func GetClientByID(ctx context.Context, clientID string) (Client, error) {
	if cacheClient[clientID] == nilClient {
		var c Client
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT client.client_id,session_id,pin_token,private_key,pin,client.name,description,assets.asset_id,speak_status,assets.icon_url,assets.symbol,information_url,welcome,host,fresh
FROM client 
LEFT JOIN assets ON client.asset_id=assets.asset_id
LEFT JOIN client_replay ON client.client_id=client_replay.client_id
LEFT JOIN client_asset_level ON client.client_id=client_asset_level.client_id
WHERE client.client_id=$1`,
			clientID).Scan(&c.ClientID, &c.SessionID, &c.PinToken, &c.PrivateKey, &c.Pin, &c.Name, &c.Description, &c.AssetID, &c.SpeakStatus, &c.IconURL, &c.Symbol, &c.InformationURL, &c.Welcome, &c.Host, &c.Amount); err != nil {
			log.Println(err)
			return c, err
		}
		cacheClient[clientID] = c
	}
	return cacheClient[clientID], nil
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

var cacheFirstClient *mixin.Client

func GetFirstClient(ctx context.Context) *mixin.Client {
	if cacheFirstClient == nil {
		c, err := GetClientList(ctx)
		if err != nil {
			return nil
		}
		cacheFirstClient = GetMixinClientByID(ctx, c[0].ClientID).Client
	}
	return cacheFirstClient
}

type MixinClient struct {
	*mixin.Client
	Secret         string
	AssetID        string
	SpeakStatus    int
	Host           string
	InformationURL string
}

var cacheHostClientMap = make(map[string]MixinClient)
var nilHostClientMap = MixinClient{}

func GetMixinClientByHost(ctx context.Context, host string) MixinClient {
	if cacheHostClientMap[host] == nilHostClientMap {
		var keystore mixin.Keystore
		var secret, assetID string
		var speakStatus int
		err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id,client_secret,session_id,pin_token,private_key,speak_status,asset_id
FROM client WHERE host=$1
`, host).Scan(&keystore.ClientID, &secret, &keystore.SessionID, &keystore.PinToken, &keystore.PrivateKey, &speakStatus, &assetID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Println(host, "...Host NOT FOUND")
			}
			return MixinClient{}
		}
		client, err := mixin.NewFromKeystore(&keystore)
		if err != nil {
			session.Logger(ctx).Println(err)
			return MixinClient{}
		}
		cacheHostClientMap[host] = MixinClient{Client: client, Secret: secret, SpeakStatus: speakStatus, AssetID: assetID}
	}
	return cacheHostClientMap[host]
}

var cacheIdClientMap = make(map[string]MixinClient)
var nilIDClientMap = MixinClient{}

func GetMixinClientByID(ctx context.Context, clientID string) MixinClient {
	if cacheIdClientMap[clientID] == nilIDClientMap {
		var keystore mixin.Keystore
		var secret, assetID, host, informationURL string
		var speakStatus int
		err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id,client_secret,session_id,pin_token,private_key,speak_status,asset_id,host,information_url
FROM client WHERE client_id=$1
`, clientID).Scan(&keystore.ClientID, &secret, &keystore.SessionID, &keystore.PinToken, &keystore.PrivateKey, &speakStatus, &assetID, &host, &informationURL)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				session.Logger(ctx).Println(err)
			}
			return MixinClient{}
		}
		client, err := mixin.NewFromKeystore(&keystore)
		if err != nil {
			session.Logger(ctx).Println(err)
			return MixinClient{}
		}
		cacheIdClientMap[clientID] = MixinClient{Client: client, Secret: secret, SpeakStatus: speakStatus, AssetID: assetID, Host: host, InformationURL: informationURL}
	}
	return cacheIdClientMap[clientID]
}

type clientInfo struct {
	*Client
	PriceUsd    decimal.Decimal `json:"price_usd,omitempty"`
	ChangeUsd   string          `json:"change_usd,omitempty"`
	TotalPeople decimal.Decimal `json:"total_people"`
	WeekPeople  decimal.Decimal `json:"week_people"`
	Activity    []*Activity     `json:"activity,omitempty"`
}

func GetClientInfoByHostOrID(ctx context.Context, host, id string) (*clientInfo, error) {
	var mixinClient MixinClient
	if id != "" {
		mixinClient = GetMixinClientByID(ctx, id)
	} else {
		mixinClient = GetMixinClientByHost(ctx, host)
	}
	client, err := GetClientByID(ctx, mixinClient.ClientID)
	if err != nil {
		return nil, err
	}
	client.SessionID = ""
	client.PinToken = ""
	client.PrivateKey = ""
	client.Pin = ""
	var c clientInfo
	c.Client = &client
	asset, err := GetAssetByID(ctx, mixinClient.Client, client.AssetID)
	if err != nil {
		return nil, err
	}
	c.PriceUsd = asset.PriceUsd
	c.ChangeUsd = asset.ChangeUsd
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

var cacheAllClient = make([]clientInfo, 0)

func GetAllClientInfo(ctx context.Context) ([]clientInfo, error) {
	if len(cacheAllClient) == 0 {
		cis := make([]clientInfo, 0)
		if err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id FROM client WHERE client_id!=ANY($1)
`, func(rows pgx.Rows) error {
			for rows.Next() {
				var clientID string
				if err := rows.Scan(&clientID); err != nil {
					return err
				}
				if ci, err := GetClientInfoByHostOrID(ctx, "", clientID); err != nil {
					return err
				} else {
					cis = append(cis, *ci)
				}
			}
			return nil
		}, config.Config.AvoidClientList); err != nil {
			return nil, err
		}
		cacheAllClient = cis
	}
	return cacheAllClient, nil
}
