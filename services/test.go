package services

import (
	"context"
	"github.com/fox-one/mixin-sdk-go"
	"log"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {

	id := mixin.UniqueConversationID("e8e8cd79-cd40-4796-8c54-3a13cfe50115", "11efbb75-e7fe-44d7-a14f-698535289310")
	log.Println(id)

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
