package oss

import (
	"github.com/alomerry/go-pusher/component/oss/kodo"
	"github.com/alomerry/go-pusher/share"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"sync"
)

var (
	Client   OSSClient
	initOnce sync.Once
)

type OSSClient struct {
	providers []string
}

type OSS interface {
	Push(filePath, key string) (string, error)
}

func InitOSS() {
	initOnce.Do(func() {
		Client = OSSClient{
			providers: cast.ToStringSlice(viper.GetStringMap("pusher")["oss-provider"]),
		}
		for _, provider := range Client.providers {
			switch provider {
			case share.OSS_PROVIDER_QI_NIU:
				kodo.InitKodo()
			default:
				panic(share.OSSNotSupport)
			}
		}
	})

}

func (o OSSClient) Push(filePath, key string) (string, error) {
	for _, provider := range o.providers {
		switch provider {
		case share.OSS_PROVIDER_QI_NIU:
			return kodo.GetKodoClient().PutFile(filePath, key)
		default:
			panic(share.OSSNotSupport)
		}
	}
	return "", nil
}
