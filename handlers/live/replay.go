package live

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

func GetLiveReplayByLiveID(ctx context.Context, u *models.ClientUser, liveID, addr string) ([]*models.LiveReplay, error) {
	lrs := make([]*models.LiveReplay, 0)
	if err := session.DB(ctx).Find(&lrs, "live_id=?", liveID).Error; err != nil {
		return nil, err
	}
	if err := session.DB(ctx).Create(&models.LivePlay{
		LiveID: liveID,
		UserID: u.UserID,
		Addr:   addr,
	}).Error; err != nil {
		tools.Println(err)
	}
	return lrs, nil
}
func HandleAudioReplay(clientID string, msg *mixin.MessageView) {
	var id, mimeType string
	var key, digest []byte
	category := common.GetPlainCategory(msg.Category)
	switch category {
	case mixin.MessageCategoryPlainText:
		msg.Data = string(tools.Base64Decode(msg.Data))
	case mixin.MessageCategoryPlainImage:
		var img mixin.ImageMessage
		if err := json.Unmarshal(tools.Base64Decode(msg.Data), &img); err != nil {
			tools.Println(err)
		}
		id = img.AttachmentID
		mimeType = img.MimeType
		if img.AttachmentMessageEncrypt != nil && len(img.Key) > 0 {
			key = img.Key
			digest = img.Digest
		}
	case mixin.MessageCategoryPlainAudio:
		var audio mixin.AudioMessage
		if err := json.Unmarshal(tools.Base64Decode(msg.Data), &audio); err != nil {
			tools.Println(err)
			return
		}
		id = audio.AttachmentID
		mimeType = audio.MimeType
		if audio.AttachmentMessageEncrypt != nil && len(audio.Key) > 0 {
			key = audio.Key
			digest = audio.Digest
		}
		b, err := getBlobFromAttachmentID(clientID, id)
		if err != nil {
			tools.Println(err)
			return
		}
		if len(key) > 0 {
			b, err = tools.DecryptAttachment(b, key, digest)
			if err != nil {
				tools.Println(err)
				return
			}
		}
		fromName := fmt.Sprintf("%s.ogg", msg.MessageID)
		toName := fmt.Sprintf("%s.mp3", msg.MessageID)
		t, err := os.Create(fromName)
		if err != nil {
			tools.Println(err)
			return
		}
		defer t.Close()
		if _, err := t.Write(b); err != nil {
			tools.Println(err)
			return
		}
		cmd := exec.Command("ffmpeg", "-i", fromName, "-acodec", "libmp3lame", "-ac", "2", "-ab", "160", toName)
		if err := cmd.Start(); err != nil {
			tools.Println(err)
			return
		}
		if err := cmd.Wait(); err != nil {
			tools.Println(err)
			return
		}

		if err := os.Remove(fromName); err != nil {
			tools.Println(err)
			return
		}
		if err := UploadFileToQiniu(toName, "live-replay/"+msg.MessageID); err != nil {
			tools.Println(err)
			return
		}
		if err := os.Remove(toName); err != nil {
			tools.Println(err)
			return
		}
		msg.Data = msg.MessageID
	case mixin.MessageCategoryPlainVideo:
		var video mixin.VideoMessage
		if err := json.Unmarshal(tools.Base64Decode(msg.Data), &video); err != nil {
			tools.Println(err)
		}
		id = video.AttachmentID
		mimeType = video.MimeType
		if video.AttachmentMessageEncrypt != nil && len(video.Key) > 0 {
			key = video.Key
			digest = video.Digest
		}
	}
	if id != "" && mimeType != "" &&
		(category != mixin.MessageCategoryPlainAudio) {
		b, err := getBlobFromAttachmentID(clientID, id)
		if err != nil {
			tools.Println(err)
		}
		if len(key) > 0 {
			b, err = tools.DecryptAttachment(b, key, digest)
			if err != nil {
				tools.Println(err)
				return
			}
		}
		err = UploadToQiniu(b, mimeType, "live-replay/"+msg.MessageID)
		if err != nil {
			tools.Println(err)
		}
		msg.Data = msg.MessageID
	}

	if err := session.DB(models.Ctx).Save(&models.LiveReplay{
		MessageID: msg.MessageID,
		ClientID:  clientID,
		Category:  category,
		Data:      msg.Data,
		CreatedAt: msg.CreatedAt,
	}); err != nil {
		tools.Println(err)
	}
}

func UploadFileToQiniu(name string, key string) error {
	formUploader, upToken := getQiniuUploader(key)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	return formUploader.PutFile(context.Background(), &ret, upToken, key, name, &putExtra)
}

func UploadToQiniu(data []byte, mimeType, key string) error {
	formUploader, upToken := getQiniuUploader(key)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		MimeType: mimeType,
	}
	dataLen := int64(len(data))
	return formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), dataLen, &putExtra)
}

func getQiniuUploader(key string) (*storage.FormUploader, string) {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", config.Config.Qiniu.Bucket, key),
	}
	mac := qbox.NewMac(config.Config.Qiniu.AccessKey, config.Config.Qiniu.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	switch config.Config.Qiniu.Region {
	case "huadong":
		cfg.Zone = &storage.ZoneHuadong
	case "huabei":
		cfg.Zone = &storage.ZoneHuabei
	case "huanan":
		cfg.Zone = &storage.ZoneHuanan
	case "beimei":
		cfg.Zone = &storage.ZoneBeimei
	case "xinjiapo":
		cfg.Zone = &storage.ZoneXinjiapo
	case "fogcneast":
		cfg.Zone = &storage.ZoneFogCnEast1
	}
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	return storage.NewFormUploader(&cfg), upToken
}

func getBlobFromAttachmentID(clientID, id string) ([]byte, error) {
	client, err := common.GetMixinClientByIDOrHost(models.Ctx, clientID)
	if err != nil {
		return nil, err
	}
	a, err := client.ShowAttachment(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return session.Api(context.Background()).RawGet(a.ViewURL), nil
}
