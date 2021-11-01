package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/tools"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	c := tools.GetRandomInvitedCode()
	log.Println(c)
	return nil
}
