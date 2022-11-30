package pusher

import (
	"errors"
	"github.com/alomerry/OSSPusher/pusher/kodo"
	"github.com/spf13/viper"
)

const (
	QI_NIU = "qiniu"
)

var Client = OSSClient{}

type OSSClient struct{}

type OSS interface {
	Push(filePath, key string) (string, error)
}

func (OSSClient) Push(filePath, key string) (string, error) {
	providers := viper.GetStringSlice("oss-provider")
	for _, provider := range providers {
		switch provider {
		case QI_NIU:
			return kodo.Push(filePath, key)
		}
	}
	return "", errors.New("provider not support")
}
