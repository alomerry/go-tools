package kodo

import (
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const (
	ridHuadong          = "z0"            // 华东
	ridHuadongZheJiang2 = "cn-east-2"     // 华东浙江 2 区
	ridHuabei           = "z1"            // 华北
	ridHuanan           = "z2"            // 华南
	ridNorthAmerica     = "na0"           // 北美
	ridSingapore        = "as0"           // 新加坡
	ridFogCnEast1       = "fog-cn-east-1" // 亚太首尔 1 区
)

var (
	kodo Kodo
)

type Kodo struct {
	accessKey string
	secretKey string
	bucket    string
	region    string

	client *qbox.Mac
}

func InitKodo() {
	kodo = Kodo{
		accessKey: cast.ToString(viper.GetStringMap("oss-qiniu")["access-key"]),
		secretKey: cast.ToString(viper.GetStringMap("oss-qiniu")["sercet-key"]),
		bucket:    cast.ToString(viper.GetStringMap("oss-qiniu")["bucket"]),
	}
}

func GetKodoClient() Kodo {
	return kodo
}

func (k Kodo) getClient() *qbox.Mac {
	if k.client != nil {
		return k.client
	}
	return qbox.NewMac(k.accessKey, k.secretKey)
}

func (k Kodo) PutFile(filePath, ossKey string) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", k.bucket, ossKey), // 覆盖写入
		ReturnBody: `{"key":"$(key)"}`,
	}
	upToken := putPolicy.UploadToken(k.getClient())

	region, _ := storage.GetRegionByID(storage.RegionID(k.region))
	cfg := storage.Config{
		Region: &region,
	}
	formUploader := storage.NewFormUploader(&cfg)
	ret := PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{},
	}
	err := formUploader.PutFile(context.Background(), &ret, upToken, ossKey, filePath, &putExtra)
	if err != nil {
		return "", err
	}
	return ret.Key, nil
}
