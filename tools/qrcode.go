package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func CheckQRCode(data []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	qrReader := qrcode.NewQRCodeReader()
	result, _ := qrReader.Decode(bmp, nil)
	if result == nil {
		return "", nil
	}
	url := result.GetText()
	if len(url) > 0 {
		return url, nil
	}
	return "", nil
}

func MessageQRFilter(ctx context.Context, client *mixin.Client, message *mixin.MessageView) (string, error) {
	var a mixin.Attachment
	src, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(src, &a)
	if err != nil {
		session.Logger(ctx).Println("validateMessage ERROR: %+v", err)
		return "", err
	}
	attachment, err := client.ShowAttachment(ctx, a.AttachmentID)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodGet, attachment.ViewURL, nil)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if url, err := CheckQRCode(data); err != nil {
		return "", err
	} else {
		return url, nil
	}
}
