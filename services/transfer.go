package services

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/models"
)

type TransferService struct{}

func (service *TransferService) Run(ctx context.Context) error {
	for {
		models.HandleTransfer(ctx)
		time.Sleep(5 * time.Second)
	}
}
