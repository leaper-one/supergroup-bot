package services

import (
	"context"
	"encoding/csv"
	"log"
	"os"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/shopspring/decimal"
)

type AirdropService struct{}

func (service *AirdropService) Run(ctx context.Context) error {
	clientID := "0baee1a4-2aff-4e53-a227-5c23ca28bfac"
	assetID := "eea900a8-b327-488c-8d8d-1428702fe240"
	f, err := os.Open("test.csv")
	if err != nil {
		log.Println("test.csv open fail...")
		return err
	}
	ReadCsv := csv.NewReader(f)
	ReadAll, err := ReadCsv.ReadAll()
	for _, line := range ReadAll {
		if line[5] == "Your Mixin ID" {
			u, err := models.GetUserByIdentityNumber(ctx, line[7])
			if err != nil {
				log.Println(err, line[7])
				continue
			}
			if line[4] == "Valid" {
				if err := models.CreateAirdrop(ctx, &models.Airdrop{
					AirdropID: clientID,
					ClientID:  clientID,
					UserID:    u.UserID,
					AssetID:   assetID,
					Amount:    decimal.NewFromFloat(0.1),
				}); err != nil {
					log.Println(err)
				}
				log.Println("addSuccess...", line[7])
			} else if line[4] == "Invalid" {
				log.Println("InvalidUser", line[7])
				_ = models.AddBlockUser(ctx, u.UserID)
			}
		}
	}
	return nil
}
