package services

import (
	"context"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"io/ioutil"
	"log"
	"time"
)

type UpdateActivityService struct{}

type Activity struct {
	ActivityIndex int    `json:"activity_index,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	Status        int    `json:"status,omitempty"`
	ImgURL        string `json:"img_url,omitempty"`
	ExpireImgURL  string `json:"expire_img_url,omitempty"`
	Action        string `json:"action,omitempty"`

	StartAt  string `json:"start_at,omitempty"`
	ExpireAt string `json:"expire_at,omitempty"`
}

func (service *UpdateActivityService) Run(ctx context.Context) error {
	var err error
	data, err := ioutil.ReadFile("activity.json")
	if err != nil {
		log.Println("activity.json open fail...")
		return err
	}
	var as []*Activity
	err = json.Unmarshal(data, &as)
	if err != nil {
		return err
	}
	for _, a := range as {
		startAt, err := time.Parse("2006-1-2 15:4:5", a.StartAt)
		if err != nil {
			return err
		}
		expireAt, err := time.Parse("2006-1-2 15:4:5", a.ExpireAt)
		if err != nil {
			return err
		}
		if err := models.UpdateActivity(ctx, &models.Activity{
			ActivityIndex: a.ActivityIndex,
			ClientID:      a.ClientID,
			Status:        a.Status,
			ImgURL:        a.ImgURL,
			ExpireImgURL:  a.ExpireImgURL,
			Action:        a.Action,
			StartAt:       startAt,
			ExpireAt:      expireAt,
		}); err != nil {
			session.Logger(ctx).Println(err)
		}
	}
	return nil
}
