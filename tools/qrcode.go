package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/tuotoo/qrcode"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func CheckQRCode(data []byte) (bool, error) {
	qrmatrix, err := qrcode.Decode(bytes.NewReader(data))
	if err != nil {
		if err.Error() == "not found error correction level and mask" {
			return true, nil
		} else if err.Error() == "lost Position Detection Pattern" {
			return false, nil
		}
		return false, err
	}
	if len(qrmatrix.Content) > 0 {
		return true, nil
	}
	return false, nil
}

func MessageQRFilter(ctx context.Context, client *mixin.Client, message *mixin.MessageView) (bool, error) {
	var a mixin.Attachment
	src, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(src, &a)
	if err != nil {
		session.Logger(ctx).Println("validateMessage ERROR: %+v", err)
		return false, err
	}
	attachment, err := client.ShowAttachment(ctx, a.AttachmentID)
	if err != nil {
		return false, err
	}
	log.Println(attachment.ViewURL)

	req, err := http.NewRequest(http.MethodGet, attachment.ViewURL, nil)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return false, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if hasURL, err := CheckQRCode(data); err != nil {
		return false, err
	} else {
		return hasURL, nil
	}
}
