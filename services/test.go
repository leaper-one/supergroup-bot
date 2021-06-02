package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/models"
	"log"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	status, err := models.GetClientUserStatusByClientIDAndUserID(ctx, "f6deb534-13bd-45f0-9b34-0d618827f500", "105f6e8b-d249-4b4d-9beb-e03cefaebc37")
	if err != nil {
		log.Println(err)
	}
	log.Println(status)
	return nil
}

//func decodeQRcode(fi io.Reader) (paymentCodeUrl *gozxing.Result) {
//	img, _, err := image.Decode(fi)
//	if err != nil {
//		fmt.Println(err)
//	}
//	// prepare BinaryBitmap
//	bmp, _ := gozxing.NewBinaryBitmapFromImage(img)
//	// decode image
//	qrReader := qrcode.NewQRCodeReader()
//	result, err := qrReader.Decode(bmp, nil)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	return result
//}
