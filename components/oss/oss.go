package oss

import (
	"errors"

	innerKodo "github.com/alomerry/go-tools/components/oss/kodo"
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
	case meta.ClientKodo:
		return innerKodo.NewKodoClient(cfg)
	default:
		return nil, ErrUnsupportedClientType
	}
}
