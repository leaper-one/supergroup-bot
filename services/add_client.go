package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/clients"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AddClientService struct{}

type clientInfo struct {
	Client      models.Client           `json:"client"`
	Level       models.ClientAssetLevel `json:"level"`
	Replay      models.ClientReplay     `json:"replay"`
	ManagerList []string                `json:"manager_list"`
}

func (service *AddClientService) Run(ctx context.Context) error {
	_, err := addClient(ctx)
	return err
}

func addClient(ctx context.Context) (*clientInfo, error) {
	var err error
	data, err := ioutil.ReadFile("client.json")
	if err != nil {
		log.Println("client.json open fail...")
		return nil, err
	}
	var client clientInfo
	err = json.Unmarshal(data, &client)
	if err != nil {
		return nil, err
	}
	client.Level.ClientID = client.Client.ClientID
	client.Replay.ClientID = client.Client.ClientID

	log.Println(client.Client.PrivateKey)
	c, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   client.Client.ClientID,
		SessionID:  client.Client.SessionID,
		PrivateKey: client.Client.PrivateKey,
		PinToken:   client.Client.PinToken,
	})
	if err != nil {
		log.Println("keystore is err...", err)
		return &client, err
	}

	m, err := c.UserMe(ctx)
	c.FavoriteApp(ctx, config.Config.LuckCoinAppID)
	client.Client.OwnerID = m.App.CreatorID
	client.Client.IdentityNumber = m.IdentityNumber
	if err := clients.InitClientMemberAuth(ctx, client.Client.ClientID); err != nil {
		log.Println("init client member auth error...", err)
		return &client, err
	}
	if err != nil {
		log.Println("user me is err...", err)
		return &client, err
	}
	client.Client.IconURL = m.AvatarURL
	client.Client.Name = m.FullName
	client.Client.SpeakStatus = common.ClientSpeckStatusClose
	if err := updateUserToManager(ctx, client.Client.ClientID, m.App.CreatorID); err != nil {
		log.Println("update manager error...", err)
	}
	if err = updateClient(ctx, client.Client); err != nil {
		log.Println(err)
	} else {
		log.Println("client update success...")
		if client.Client.AssetID != "" {
			common.GetAssetByID(ctx, c, client.Client.AssetID)
		}
	}
	if err = updateClientReplay(ctx, client.Replay); err != nil {
		log.Println(err)
	} else {
		log.Println("client_replay update success...")
	}
	if err = updateClientAssetLevel(ctx, client.Level, client.Client.AssetID); err != nil {
		log.Println(err)
	} else {
		log.Println("level update success")
	}
	for _, s := range client.ManagerList {
		if err := updateManagerList(ctx, client.Client.ClientID, s); err != nil {
			log.Println("update manager error...", err)
		}
	}
	return &client, nil
}

func updateClient(ctx context.Context, client models.Client) error {
	if !checkClientField(&client) {
		return nil
	}
	go session.Redis(ctx).QDel(ctx, "client:"+client.ClientID)
	if err := session.DB(ctx).Save(&client).Error; err != nil {
		return err
	}
	return nil
}
func updateClientReplay(ctx context.Context, cr models.ClientReplay) error {
	if cr.ClientID == "" {
		log.Println("client_replay client_id 不能为空")
		return nil
	}
	if err := session.DB(ctx).Save(&cr).Error; err != nil {
		return err
	}
	return nil
}

func updateClientAssetLevel(ctx context.Context, l models.ClientAssetLevel, assetID string) error {
	if l.ClientID == "" {
		log.Println("level client_id 不能为空")
		return nil
	}
	if assetID == "" {
		l.Fresh = decimal.NewFromInt(100)
		l.Senior = decimal.NewFromInt(2000)
		l.Large = decimal.NewFromInt(2000)
		l.FreshAmount = decimal.NewFromInt(1)
		l.LargeAmount = decimal.NewFromInt(10)
	}
	if err := session.DB(ctx).Save(&l).Error; err != nil {
		return err
	}
	return nil
}

func checkClientField(client *models.Client) bool {
	if client.ClientID == "" {
		return tips("client_id 不能为空")
	}
	if client.SessionID == "" {
		return tips("session_id 不能为空")
	}
	if client.PinToken == "" {
		return tips("pin_token 不能为空")
	}
	if client.PrivateKey == "" {
		return tips("private_key 不能为空")
	}
	return true
}

func tips(msg string) bool {
	log.Println(msg)
	return false
}

func updateUserToManager(ctx context.Context, clientID string, userID string) error {
	_, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := session.DB(ctx).Create(&models.ClientUser{
				ClientID:     clientID,
				UserID:       userID,
				Priority:     models.ClientUserPriorityHigh,
				Status:       models.ClientUserStatusAdmin,
				IsReceived:   true,
				IsNoticeJoin: true,
				CreatedAt:    time.Now(),
			}).Error; err != nil {
				return err
			}

		}
	} else {
		if err := session.DB(ctx).Table("client_users").
			Where("client_id=? AND user_id=?", clientID, userID).
			Updates(map[string]interface{}{
				"priority": models.ClientUserPriorityHigh,
				"status":   models.ClientUserStatusAdmin,
			}).Error; err != nil {
			return err
		}
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return nil
}

func updateManagerList(ctx context.Context, clientID string, id string) error {
	u, err := common.SearchUser(ctx, clientID, id)
	if err != nil {
		return err
	}
	if err := session.DB(ctx).Create(&models.User{
		UserID:         u.UserID,
		IdentityNumber: u.IdentityNumber,
		AvatarURL:      u.AvatarURL,
		FullName:       u.FullName,
		IsScam:         u.IsScam,
	}).Error; err != nil {
		return err
	}
	if err := updateUserToManager(ctx, clientID, u.UserID); err != nil {
		return err
	}
	return nil
}
