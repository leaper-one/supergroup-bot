package live

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/fox-one/mixin-sdk-go"
)

func UploadLiveImgToMixinStatistics(ctx context.Context, u *models.ClientUser, r *http.Request) (string, error) {
	if err := r.ParseMultipartForm(10 * 1024); err != nil {
		return "", err
	}
	image := r.MultipartForm.File["file"]
	file, err := image[0].Open()
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
		//return
	}
	client, err := common.GetMixinClientByIDOrHost(ctx, u.ClientID)
	if err != nil {
		return "", err
	}
	a, err := client.CreateAttachment(ctx)
	if err != nil {
		return "", err
	}
	if err := mixin.UploadAttachment(ctx, a, data); err != nil {
		return "", err
	}
	return a.ViewURL, nil
}
