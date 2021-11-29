package services

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	models.HandleStatictis(ctx)
	return nil
}
