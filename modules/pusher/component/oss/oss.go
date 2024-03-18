package oss

import (
	"sync"

	"github.com/alomerry/go-tools/modules/pusher/component/oss/kodo"
	"github.com/alomerry/go-tools/modules/pusher/share"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
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
	Delete(keys []string) error
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

func (o OSSClient) Delete(keys []string) error {
	for _, provider := range o.providers {
		switch provider {
		case share.OSS_PROVIDER_QI_NIU:
			return kodo.GetKodoClient().DeleteFiles(keys)
		default:
			panic(share.OSSNotSupport)
		}
	}
	return nil
}
