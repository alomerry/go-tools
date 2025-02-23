package oss

import (
	"errors"

	"github.com/alomerry/go-tools/components/oss/meta"
	innerMinio "github.com/alomerry/go-tools/components/oss/minio"
)

var (
	ErrUnsupportedClientType = errors.New("unsupported client type")
)

func NewClient(cfg meta.Config) (meta.OSSClient, error) {
	switch cfg.Type {
	case meta.ClientTypeMinio:
		return innerMinio.NewMinioClient(cfg)
	default:
		return nil, ErrUnsupportedClientType
	}
}
