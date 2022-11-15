package models

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

const client_user_proxy_DDL = `
CREATE TABLE IF NOT EXISTS client_user_proxy (
  client_id         VARCHAR(36) NOT NULL,
  proxy_user_id     VARCHAR(36) NOT NULL,

	user_id 					VARCHAR(36) NOT NULL,	
	full_name 				VARCHAR(255) NOT NULL,
	session_id 				VARCHAR(36) NOT NULL,
	pin_token 				VARCHAR NOT NULL,
	private_key 			VARCHAR NOT NULL,
	
	status            SMALLINT NOT NULL DEFAULT 1, -- 1: inactive, 2: active
  created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (client_id, proxy_user_id)
);
CREATE INDEX IF NOT EXISTS client_user_proxy_user_idx ON client_user_proxy (client_id, user_id);
`

type ClientUserProxy struct {
	ClientID    string    `json:"client_id,omitempty"`
	ProxyUserID string    `json:"proxy_user_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	FullName    string    `json:"full_name,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	PinToken    string    `json:"pin_token,omitempty"`
	PrivateKey  string    `json:"private_key,omitempty"`
	Status      int       `json:"status,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

const (
	ClientUserProxyStatusInactive = 1
	ClientUserProxyStatusActive   = 2
)

func UpdateClientUserProxy(ctx context.Context, u *ClientUser, isProxy bool, fullName string, avatarURL string) error {
	var status int
	if isProxy {
		status = ClientUserProxyStatusActive
	} else {
		status = ClientUserProxyStatusInactive
	}
	up, err := getClientUserProxyByProxyID(ctx, u.ClientID, u.UserID)
	if err != nil {
		return err
	}
	if fullName != up.FullName {
		c, err := mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   up.UserID,
			SessionID:  up.SessionID,
			PrivateKey: up.PrivateKey,
		})
		if err != nil {
			return err
		}
		avatarBase64, err := getBase64AvatarByURL(avatarURL)
		if err != nil {
			return err
		}
		if _, err := c.ModifyProfile(ctx, fullName, avatarBase64); err != nil {
			return err
		}
	}

	_, err = session.Database(ctx).Exec(ctx, `
UPDATE client_user_proxy
SET status = $1, full_name = $2
WHERE client_id = $3 AND proxy_user_id = $4
	`, status, fullName, u.ClientID, u.UserID)
	return err
}

func getClientUserProxyByProxyID(ctx context.Context, clientID, userID string) (*ClientUserProxy, error) {
	var cup ClientUserProxy
	err := session.Database(ctx).QueryRow(ctx, `
SELECT user_id, status, full_name, session_id, pin_token, private_key
FROM client_user_proxy
WHERE client_id = $1 AND proxy_user_id = $2
	`, clientID, userID).Scan(&cup.UserID, &cup.Status, &cup.FullName, &cup.SessionID, &cup.PinToken, &cup.PrivateKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return newProxyUser(ctx, clientID, userID)
	}
	return &cup, err
}

func checkAndReplaceProxyUser(ctx context.Context, clientID string, userID *string) {
	var cup ClientUserProxy
	err := session.Database(ctx).QueryRow(ctx, `
SELECT proxy_user_id, status
FROM client_user_proxy
WHERE client_id = $1 AND user_id = $2
	`, clientID, userID).Scan(&cup.ProxyUserID, &cup.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return
	}
	if cup.Status == ClientUserProxyStatusActive {
		*userID = cup.ProxyUserID
	}
}

func newProxyUser(ctx context.Context, clientID, userID string) (*ClientUserProxy, error) {
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return nil, err
	}
	_u, err := getUserByID(ctx, userID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return nil, err
	}
	_, keystore, err := client.CreateUser(ctx, mixin.GenerateEd25519Key(), _u.FullName)
	if err != nil {
		return nil, err
	}
	uc, err := mixin.NewFromKeystore(keystore)
	if err != nil {
		return nil, err
	}
	base64Avatar, err := getBase64AvatarByURL(_u.AvatarURL)
	if err != nil {
		return nil, err
	}
	if _, err := uc.ModifyProfile(ctx, _u.FullName, base64Avatar); err != nil {
		return nil, err
	}
	cup := &ClientUserProxy{
		ClientID:    clientID,
		ProxyUserID: userID,
		UserID:      keystore.ClientID,
		FullName:    _u.FullName,
		SessionID:   keystore.SessionID,
		PinToken:    keystore.PinToken,
		PrivateKey:  keystore.PrivateKey,
		Status:      ClientUserProxyStatusInactive,
	}
	if err := createClientUserProxy(ctx, cup); err != nil {
		session.Logger(ctx).Println(err)
		return getClientUserProxyByProxyID(ctx, clientID, userID)
	}
	return cup, nil
}

func createClientUserProxy(ctx context.Context, u *ClientUserProxy) error {
	query := durable.InsertQuery("client_user_proxy", "client_id, proxy_user_id, user_id, full_name, session_id, pin_token, private_key, status")
	_, err := session.Database(ctx).Exec(ctx, query, u.ClientID, u.ProxyUserID, u.UserID, u.FullName, u.SessionID, u.PinToken, u.PrivateKey, u.Status)
	return err
}

func getBase64AvatarByURL(url string) (string, error) {
	if url == "" || url == DefaultAvatar {
		return "", nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sourceString := base64.StdEncoding.EncodeToString(body)
	return sourceString, nil
}

func DailyUpdateProxyUserProfile() {
	for {
		updateAllProxyUserProfile(_ctx)
		time.Sleep(time.Hour * 24)
	}
}

func updateAllProxyUserProfile(ctx context.Context) error {
	clients, err := GetAllClient(ctx)
	if err != nil {
		return err
	}

	for _, clientID := range clients {
		status := GetClientProxy(ctx, clientID)
		if status == ClientProxyStatusOff {
			continue
		}
		// 1. 拿到所有的 用户
		_users, err := getClientUserByClientID(ctx, clientID, 0)
		if err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
		for _, userID := range _users {
			u, err := getUserByID(ctx, userID)
			if err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
			if err := UpdateClientUserProxy(ctx, &ClientUser{ClientID: clientID, UserID: userID}, true, u.FullName, u.AvatarURL); err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
		}
	}

	return nil
}
