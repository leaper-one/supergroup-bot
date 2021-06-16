package services

import (
	"context"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/models"
	"io/ioutil"
	"log"
)

type AddClientService struct{}

type clientInfo struct {
	Client *models.Client           `json:"client"`
	Level  *models.ClientAssetLevel `json:"level"`
	Replay *models.ClientReplay     `json:"replay"`
}

func (service *AddClientService) Run(ctx context.Context) error {
	var err error
	data, err := ioutil.ReadFile("client.json")
	if err != nil {
		log.Println("client.json open fail...")
		return err
	}
	var client clientInfo
	err = json.Unmarshal(data, &client)
	client.Level.ClientID = client.Client.ClientID
	client.Replay.ClientID = client.Client.ClientID
	if err != nil {
		return err
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

	return nil
}

func updateClient(ctx context.Context, client *models.Client) error {
	if !checkClientField(client) {
		return nil
	}
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
	if client.AssetID == "" {
		return tips("asset_id 不能为空")
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
