package services

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

type AddClientService struct{}

type clientInfo struct {
	Client      *models.Client           `json:"client"`
	Level       *models.ClientAssetLevel `json:"level"`
	Replay      *models.ClientReplay     `json:"replay"`
	ManagerList []string                 `json:"manager_list"`
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
	if err != nil {
		log.Println("user me is err...", err)
		return &client, err
	}
	if client.Client.AssetID == "" {
		client.Client.IconURL = m.AvatarURL
	}
	if client.Level.Fresh.IsZero() {
		client.Client.SpeakStatus = models.ClientSpeckStatusClose
	} else {
		client.Client.SpeakStatus = models.ClientSpeckStatusOpen
	}

	if err := updateUserToManager(ctx, client.Client.ClientID, m.App.CreatorID); err != nil {
		log.Println("update manager error...", err)
	}
	if err = updateClient(ctx, client.Client); err != nil {
		log.Println(err)
	} else {
		log.Println("client update success...")
		if client.Client.AssetID != "" {
			models.GetAssetByID(ctx, models.GetFirstClient(ctx), client.Client.AssetID)
		}
	}
	if err = updateClientReplay(ctx, client.Replay); err != nil {
		log.Println(err)
	} else {
		log.Println("client_replay update success...")
	}
	if err = updateClientAssetLevel(ctx, client.Level); err != nil {
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

func updateClient(ctx context.Context, client *models.Client) error {
	if !checkClientField(client) {
		return nil
	}
	if err := models.UpdateClient(ctx, client); err != nil {
		return err
	}
	return nil
}
func updateClientReplay(ctx context.Context, cr *models.ClientReplay) error {
	if cr.ClientID == "" {
		log.Println("client_replay client_id 不能为空")
		return nil
	}
	if err := models.UpdateClientReplay(ctx, cr); err != nil {
		return err
	}
	return nil
}

func updateClientAssetLevel(ctx context.Context, l *models.ClientAssetLevel) error {
	if l == nil {
		log.Println("未发现 level...")
		return nil
	}
	if l.ClientID == "" {
		log.Println("level client_id 不能为空")
		return nil
	}
	if err := models.UpdateClientAssetLevel(ctx, l); err != nil {
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
	if client.Pin == "" {
		return tips("pin 不能为空")
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
	_, err := models.GetClientUserByClientIDAndUserID(ctx, clientID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			query := durable.InsertQuery("client_users", "client_id,user_id,access_token,priority,status")
			_, err := session.Database(ctx).Exec(ctx, query, clientID, userID, "", models.ClientUserPriorityHigh, models.ClientUserStatusAdmin)
			if err != nil {
				return err
			}
		}
	} else {
		_, err := session.Database(ctx).Exec(ctx, `
UPDATE client_users SET priority=1,status=9 WHERE client_id=$1 AND user_id=$2
`, clientID, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateManagerList(ctx context.Context, clientID string, id string) error {
	u, err := models.SearchUser(ctx, id)
	if err != nil {
		return err
	}
	if err := models.WriteUser(ctx, &models.User{
		UserID:         u.UserID,
		IdentityNumber: u.IdentityNumber,
		AvatarURL:      u.AvatarURL,
		FullName:       u.FullName,
	}); err != nil {
		return err
	}
	if err := updateUserToManager(ctx, clientID, u.UserID); err != nil {
		return err
	}
	return nil
}
