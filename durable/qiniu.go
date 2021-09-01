package durable

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Qiniu struct {
}

func (q *Qiniu) init() {
	if config.Config.Qiniu.AccessKey == "" {
		return
	}

	putPolicy := storage.PutPolicy{
		Scope: config.Config.Qiniu.Bucket,
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
	if cfg.Zone == nil {
		log.Panic("qiniu.region is invalid...")
	}

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
