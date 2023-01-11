package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	var cus []*models.ClientUser
	log.Println(1)
	session.DB(ctx).Debug().Find(&cus, "priority=4")
	log.Println(2)
	log.Println(len(cus))
	for i, u := range cus {
		log.Println(i, len(cus))
		if u.Status != models.ClientUserStatusBlock {
			if u.Status == models.ClientUserStatusAdmin || u.Status == models.ClientUserStatusGuest {
				session.DB(ctx).Model(u).Update("priority", models.ClientUserPriorityHigh)
			} else {
				session.DB(ctx).Model(u).Update("priority", models.ClientUserPriorityLow)
			}
		}
	}
	return nil
}
