package services

import (
	"context"
	"log"

	"github.com/fox-one/mixin-sdk-go"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	asset, _ := mixin.HashFromString("b9f49cf777dc4d03bc54cd1367eebca319f8603ea1ce18910d09e2c540c630d8")
	input1, _ := mixin.HashFromString("8bef65dc4eb1705600eeefab066bbd66d710647aa269581eec42510b3c5efc18")
	output1, _ := mixin.KeyFromString("2d1a9f6b4fd35514b10d2834c489d393593bd86de39d886a08f92d728bebc1dc")
	output1key, _ := mixin.KeyFromString("3d2d0a041ce370e17e47c3c7137116053d6040be0da190aef1d61bf0f110d9f1")
	output2, _ := mixin.KeyFromString("ee22b0355a0b1c12d97304757d7dbc7d42e89008fcf508f3b4ce30412d77359f")
	output2key, _ := mixin.KeyFromString("10522fd23e3f7f040152c5ec7c056f9a1a85363d004f87ac56d16c7e87a450a6")
	output2key2, _ := mixin.KeyFromString("b5861f81c2cef2967fa133796ab81fdebfcd331fc630f34bba01a28e3b08363d")
	output2key3, _ := mixin.KeyFromString("bcc3ae6898f9b6813c5afee04f010a9fcd568805ad9bc0c6618f6a350d759d32")
	tx := mixin.Transaction{
		Version: 2,
		Asset:   asset,
		Extra:   []byte("multisig test"),
		Inputs: []*mixin.Input{
			{
				Hash:  &input1,
				Index: 0,
			},
		},
		Outputs: []*mixin.Output{
			{
				Mask:   output1,
				Keys:   []mixin.Key{output1key},
				Amount: mixin.NewInteger(0),
			},
			{
				Mask:   output2,
				Keys:   []mixin.Key{output2key, output2key2, output2key3},
				Amount: mixin.NewInteger(0),
			},
		},
	}
	tx1, err := tx.DumpTransaction()
	if err != nil {
		return err
	}
	log.Println(tx1)

	return nil
}
