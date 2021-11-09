package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/models"
	"log"

	"github.com/MixinNetwork/supergroup/tools"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	c := tools.GetRandomInvitedCode()
	log.Println(c)
	models.HandleInvitationOnceReward(ctx)
	return nil
}
