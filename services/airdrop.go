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
	clientID := "29a01a4f-0a42-402f-869d-34e148589982"
	assetID := "965e5c6e-434c-3fa9-b780-c50f43cd955c"
	// client, err := models.GetClientByID(ctx, clientID)
	// if err != nil {
	// 	return err
	// }
	f, err := os.Open("test.csv")
	if err != nil {
		log.Println("test.csv open fail...")
		return err
	}

	//创建csv读取接口实例
	ReadCsv := csv.NewReader(f)

	//获取一行内容，一般为第一行内容
	// read, _ := ReadCsv.Read() //返回切片类型：[chen  hai wei]
	// tools.PrintJson(read)

	//读取所有内容
	ReadAll, err := ReadCsv.ReadAll() //返回切片类型：[[s s ds] [a a a]]
	for _, line := range ReadAll {
		if line[5] == "Your Mixin ID" {
			u, err := models.GetUserByIdentityNumber(ctx, line[7])
			if err != nil {
				log.Println(err, line[7])
				continue
			}

			if err := models.CreateAirdrop(ctx, &models.Airdrop{
				AirdropID: clientID,
				ClientID:  clientID,
				UserID:    u.UserID,
				AssetID:   assetID,
				Amount:    decimal.NewFromInt(1),
			}); err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}
