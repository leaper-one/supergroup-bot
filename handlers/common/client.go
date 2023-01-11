package common

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
)

const (
	ClientSpeckStatusOpen  = 1 // 持仓发言打开，
	ClientSpeckStatusClose = 2 // 持仓发言关闭

	ClientPayStatusOpen = 1 // 入群开启，
)

func GetClientByIDOrHost(ctx context.Context, clientIDorHost string) (models.Client, error) {
	var c models.Client
	key := "client:" + clientIDorHost
	if err := session.Redis(ctx).StructScan(ctx, key, &c); err != nil {
		if errors.Is(err, redis.Nil) {
			return CacheClient(ctx, clientIDorHost)
		}
		tools.Println(err)
		return c, err
	}
	return c, nil
}

func CacheClient(ctx context.Context, clientIDOrHost string) (models.Client, error) {
	var c models.Client
	if err := session.DB(ctx).First(&c, "client_id=? OR host=?", clientIDOrHost, clientIDOrHost).Error; err != nil {
		return c, err
	}
	var cr models.ClientReplay
	if err := session.DB(ctx).First(&cr, "client_id=?", c.ClientID).Error; err != nil {
		return c, err
	}
	c.JoinMsg = cr.JoinMsg
	c.Welcome = cr.Welcome
	key1 := "client:" + c.ClientID
	key2 := "client:" + c.Host
	if err := session.Redis(ctx).StructSet(ctx, key1, c); err != nil {
		tools.Println(err)
		return c, nil
	}
	if err := session.Redis(ctx).StructSet(ctx, key2, c); err != nil {
		tools.Println(err)
		return c, nil
	}
	return c, nil
}

type MixinClient struct {
	*mixin.Client
	C models.Client
}

var cacheClientMap = tools.NewMutex()

func GetMixinClientByIDOrHost(ctx context.Context, clientIDOrHost string) (*MixinClient, error) {
	client := cacheClientMap.Read(clientIDOrHost)
	if client == nil {
		c, err := GetClientByIDOrHost(ctx, clientIDOrHost)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				tools.Println(err, clientIDOrHost)
			}
			return nil, err
		}
		client, err := mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   c.ClientID,
			SessionID:  c.SessionID,
			PinToken:   c.PinToken,
			PrivateKey: c.PrivateKey,
		})
		if err != nil {
			tools.Println(err)
			return nil, err
		}
		_client := MixinClient{
			Client: client,
			C:      c,
		}
		cacheClientMap.Write(clientIDOrHost, &_client)
		return &_client, nil
	} else {
		return client.(*MixinClient), nil
	}
}

func getMixinOAuthClientByClientUser(ctx context.Context, u *models.ClientUser) (*mixin.Client, error) {
	return mixin.NewFromOauthKeystore(&mixin.OauthKeystore{
		ClientID:   u.ClientID,
		AuthID:     u.AuthorizationID,
		Scope:      u.Scope,
		PrivateKey: u.PrivateKey,
		VerifyKey:  u.Ed25519,
	})
}

func getClientAdminOrOwner(ctx context.Context, clientId string) (*models.User, error) {
	c, err := GetClientByIDOrHost(ctx, clientId)
	if err != nil {
		return nil, err
	}
	adminId := c.AdminID
	if adminId == "" {
		adminId = c.OwnerID
	}
	return SearchUser(ctx, clientId, adminId)
}

func MuteClientOperation(muteStatus bool, clientID string) {
	ctx := models.Ctx
	if muteStatus {
		// 1. 如果是关闭
		if err := SetClientConversationStatusByIDAndStatus(ctx, clientID, models.ClientConversationStatusMute); err != nil {
			tools.Println(err)
		} else {
			DeleteDistributeMsgByClientID(ctx, clientID)
			go SendClientTextMsg(clientID, config.Text.MuteOpen, "", false)
		}

	} else {
		// 2. 如果是打开
		if err := SetClientConversationStatusByIDAndStatus(ctx, clientID, models.ClientConversationStatusNormal); err != nil {
			tools.Println(err)
		} else {
			go SendClientTextMsg(clientID, config.Text.MuteClose, "", false)
		}
	}
}
