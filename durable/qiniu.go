package durable

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Qiniu struct {
}

func (q *Qiniu) init() {
	putPolicy := storage.PutPolicy{
		Scope: config.Config.Qiniu.Bucket,
	}
	mac := qbox.NewMac(config.Config.Qiniu.AccessKey, config.Config.Qiniu.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": "github logo",
		},
	}
	data := []byte("hello, this is qiniu cloud")
	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, "", bytes.NewReader(data), dataLen, &putExtra)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.Key, ret.Hash)

}
