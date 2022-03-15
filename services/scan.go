package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/models"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	user := *new(models.MixinClient)
	log.Println(user.ClientID)
	return nil
}
