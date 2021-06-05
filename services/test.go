package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/models"
	"log"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	//status, err := models.GetClientUserStatusByClientIDAndUserID(ctx, "f6deb534-13bd-45f0-9b34-0d618827f500", "105f6e8b-d249-4b4d-9beb-e03cefaebc37")
	//if err != nil {
	//	log.Println(err)
	//}
	//log.Println(status)

	//for i := 0; i < 10; i++ {
	//	getRid(fmt.Sprintf("f6deb534-13bd-45f0-9b34-0d618827f500%d", i))
	//}
	status, err := models.GetClientUserStatusByClientIDAndUserID(ctx, "47b0b809-2bb5-4c94-becd-35fb93f5c6fe", "a3ce6c86-307a-4187-98b0-76424cbc0fbf")
	if err != nil {
		return err
	}
	log.Println(status)

	return nil
}

func getRid(uid string) {
	//conversationID := mixin.UniqueConversationID(uid, "105f6e8b-d249-4b4d-9beb-e03cefaebc37")

	//rid, _ := tools.ShardId(conversationID, uid)
	//log.Println(conversationID)
	//log.Println(rid)
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
