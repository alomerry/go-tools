package oss

import (
	"github.com/alomerry/go-pusher/component/oss/kodo"
	"github.com/alomerry/go-pusher/share"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var Client OSSClient

type OSSClient struct {
	providers []string
}

type OSS interface {
	Push(filePath, key string) (string, error)
}

func init() {
	Client = OSSClient{
		providers: cast.ToStringSlice(viper.GetStringMap("pusher")["oss-provider"]),
	}
}

func (o OSSClient) Push(filePath, key string) (string, error) {
	for _, provider := range o.providers {
		switch provider {
		case share.OSS_PROVIDER_QI_NIU:
			return kodo.GetKodoClient().PutFile(filePath, key)
		default:
			panic("not support other oss yet.")
		}
	}
	return "", nil
}
