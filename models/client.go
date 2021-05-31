package models

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"log"
	"strings"
	"time"
)

const client_DDL = `
-- 机器人信息
CREATE TABLE IF NOT EXISTS client {
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
}
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
}

const (
	ClientSpeckStatusOpen  = 1 // 持仓发言打开，
	ClientSpeckStatusClose = 2 // 持仓发言关闭
)

func UpdateClient(ctx context.Context, c *Client) error {
	if strings.HasSuffix(c.Host, "/") {
		c.Host = c.Host[:len(c.Host)-1]
	}
	query := durable.InsertQueryOrUpdate("client", "client_id", "client_secret,session_id,pin_token,private_key,pin,name,description,asset_id,host,information_url")
	_, err := session.Database(ctx).Exec(ctx, query, c.ClientID, c.ClientSecret, c.SessionID, c.PinToken, c.PrivateKey, c.Pin, c.Name, c.Description, c.AssetID, c.Host, c.InformationURL)
	return err
}

var cacheClient = make(map[string]*Client)

func GetClientByID(ctx context.Context, clientID string) (*Client, error) {
	if cacheClient[clientID] == nil {
		var c Client
		if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT client_id,session_id,pin_token,private_key,pin,client.name,description,assets.asset_id,speak_status,assets.icon_url,assets.symbol,information_url
FROM client 
LEFT JOIN assets ON client.asset_id=assets.asset_id
WHERE client_id=$1`, func(row pgx.Row) error {
			return row.Scan(&c.ClientID, &c.SessionID, &c.PinToken, &c.PrivateKey, &c.Pin, &c.Name, &c.Description, &c.AssetID, &c.SpeakStatus, &c.IconURL, &c.Symbol, &c.InformationURL)
		}, clientID); err != nil {
			log.Println(err)
			return nil, err
		}
		cacheClient[clientID] = &c
	}
	return cacheClient[clientID], nil
}

func GetClientList(ctx context.Context) ([]*Client, error) {
	clientList := make([]*Client, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id, session_id,pin_token,private_key,pin,speak_status,created_at
FROM client
WHERE client_id=ANY($1)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var c Client
			if err := rows.Scan(&c.ClientID, &c.SessionID, &c.PinToken, &c.PrivateKey, &c.Pin, &c.SpeakStatus, &c.CreatedAt); err != nil {
				return err
			}
			clientList = append(clientList, &c)
		}
		return nil
	}, config.ClientList)
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

var cacheHostClientMap = make(map[string]*MixinClient)

func GetMixinClientByHost(ctx context.Context, host string) *MixinClient {
	if cacheHostClientMap[host] == nil {
		var keystore mixin.Keystore
		var secret, assetID string
		var speakStatus int
		err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT client_id,client_secret,session_id,pin_token,private_key,speak_status,asset_id
FROM client WHERE host=$1
`, func(row pgx.Row) error {
			return row.Scan(&keystore.ClientID, &secret, &keystore.SessionID, &keystore.PinToken, &keystore.PrivateKey, &speakStatus, &assetID)
		}, host)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Println(host, "...Host NOT FOUND")
			}
			return nil
		}
		client, err := mixin.NewFromKeystore(&keystore)
		if err != nil {
			session.Logger(ctx).Println(err)
			return nil
		}
		cacheHostClientMap[host] = &MixinClient{Client: client, Secret: secret, SpeakStatus: speakStatus, AssetID: assetID}
	}
	return cacheHostClientMap[host]
}

var cacheIdClientMap = make(map[string]*MixinClient)

func GetMixinClientByID(ctx context.Context, clientID string) *MixinClient {
	if cacheIdClientMap[clientID] == nil {
		var keystore mixin.Keystore
		var secret, assetID, host, informationURL string
		var speakStatus int
		err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT client_id,client_secret,session_id,pin_token,private_key,speak_status,asset_id,host,information_url
FROM client WHERE client_id=$1
`, func(row pgx.Row) error {
			return row.Scan(&keystore.ClientID, &secret, &keystore.SessionID, &keystore.PinToken, &keystore.PrivateKey, &speakStatus, &assetID, &host, &informationURL)
		}, clientID)
		if err != nil {
			session.Logger(ctx).Println(err)
			return nil
		}
		client, err := mixin.NewFromKeystore(&keystore)
		if err != nil {
			session.Logger(ctx).Println(err)
			return nil
		}
		cacheIdClientMap[clientID] = &MixinClient{Client: client, Secret: secret, SpeakStatus: speakStatus, AssetID: assetID, Host: host, InformationURL: informationURL}
	}
	return cacheIdClientMap[clientID]
}

func GetClientInfoByHost(ctx context.Context, host string) (*Client, error) {
	mixinClient := GetMixinClientByHost(ctx, host)
	client, err := GetClientByID(ctx, mixinClient.ClientID)
	if err != nil {
		return nil, err
	}
	return &Client{
		ClientID:       client.ClientID,
		Name:           client.Name,
		Description:    client.Description,
		Host:           client.Host,
		AssetID:        client.AssetID,
		SpeakStatus:    client.SpeakStatus,
		CreatedAt:      client.CreatedAt,
		IconURL:        client.IconURL,
		Symbol:         client.Symbol,
		InformationURL: client.InformationURL,
	}, nil
}
