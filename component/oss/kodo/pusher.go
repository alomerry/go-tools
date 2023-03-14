package kodo

import (
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var kodo Kodo

func init() {
	kodo = Kodo{
		accessKey: cast.ToString(viper.GetStringMap("oss-qiniu")["access-key"]),
		secretKey: cast.ToString(viper.GetStringMap("oss-qiniu")["sercet-key"]),
		bucket:    cast.ToString(viper.GetStringMap("oss-qiniu")["bucket"]),
	}
}

type Kodo struct {
	accessKey string
	secretKey string
	bucket    string
	client    *qbox.Mac
}

func (k Kodo) getClient() *qbox.Mac {
	if k.client != nil {
		return k.client
	}
	return qbox.NewMac(k.accessKey, k.secretKey)
}

// 自定义返回值结构体
type PutRet struct {
	Key string
}

func GetKodoClient() Kodo {
	return kodo
}

// - /Users/alomerry/workspace/OSSPusher/public/.oss_pusher_hash
// - blog/.oss_pusher_hash
func (k Kodo) PutFile(filePath, ossKey string) (string, error) {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", k.bucket, ossKey), // 覆盖写入
		ReturnBody: `{"key":"$(key)"}`,
	}
	upToken := putPolicy.UploadToken(k.getClient())

	cfg := storage.Config{}
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
