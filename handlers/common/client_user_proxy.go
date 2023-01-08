package common

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	ClientUserProxyStatusInactive = 1
	ClientUserProxyStatusActive   = 2
)

func UpdateClientUserProxy(ctx context.Context, u *models.ClientUser, isProxy bool, fullName string, avatarURL string) error {
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
	return session.DB(ctx).Model(&models.ClientUserProxy{}).
		Where("client_id = ? AND proxy_user_id = ?", u.ClientID, u.UserID).
		Updates(models.ClientUserProxy{Status: status, FullName: fullName}).Error
}

func getClientUserProxyByProxyID(ctx context.Context, clientID, userID string) (*models.ClientUserProxy, error) {
	var cup models.ClientUserProxy
	if err := session.Redis(ctx).StructScan(ctx, fmt.Sprintf("client_user_proxy:%s:%s", clientID, userID), &cup); err != nil {
		if errors.Is(err, redis.Nil) {
			err = session.DB(ctx).Take(&cup, "client_id = ? AND proxy_user_id = ?", clientID, userID).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					_cup, err := newProxyUser(ctx, clientID, userID)
					if err != nil {
						return nil, err
					}
					cup = *_cup
				} else {
					return nil, err
				}
			}
			session.Redis(ctx).StructSet(ctx, fmt.Sprintf("client_user_proxy:%s:%s", clientID, userID), &cup)
		}
	}
	err := session.DB(ctx).Take(&cup, "client_id = ? AND proxy_user_id = ?", clientID, userID).Error
	return &cup, err
}

func checkAndReplaceProxyUser(ctx context.Context, clientID string, userID *string) {
	cup, err := getClientUserProxyByProxyID(ctx, clientID, *userID)
	if err != nil {
		tools.Println(err)
		return
	}
	if cup.Status == ClientUserProxyStatusActive {
		*userID = cup.ProxyUserID
	}
}

func newProxyUser(ctx context.Context, clientID, userID string) (*models.ClientUserProxy, error) {
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return nil, err
	}
	_u, err := SearchUser(ctx, clientID, userID)
	if err != nil {
		tools.Println(err)
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
	cup := models.ClientUserProxy{
		ClientID:    clientID,
		ProxyUserID: userID,
		UserID:      keystore.ClientID,
		FullName:    _u.FullName,
		SessionID:   keystore.SessionID,
		PinToken:    keystore.PinToken,
		PrivateKey:  keystore.PrivateKey,
		Status:      ClientUserProxyStatusInactive,
	}
	if err := session.DB(ctx).Create(&cup).Error; err != nil {
		tools.Println(err)
		return getClientUserProxyByProxyID(ctx, clientID, userID)
	}
	return &cup, nil
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
