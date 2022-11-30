package kodo

import (
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// 自定义返回值结构体
type PutRet struct {
	Key string
}

// - /Users/alomerry/workspace/OSSPusher/public/.oss_pusher_hash
// - blog/.oss_pusher_hash
func Push(filePath, key string) (string, error) {
	var (
		accessKey = cast.ToString(viper.GetStringMap("oss-qiniu")["access-key"])
		secretKey = cast.ToString(viper.GetStringMap("oss-qiniu")["sercet-key"])
		bucket    = cast.ToString(viper.GetStringMap("oss-qiniu")["bucket"])
		ossKey    = key
	)

	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", bucket, ossKey), // 覆盖写入
		ReturnBody: `{"key":"$(key)"}`,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)

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
