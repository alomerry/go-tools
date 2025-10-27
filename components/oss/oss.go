package oss

import (
	"errors"

	innerKodo "github.com/alomerry/go-tools/components/oss/kodo"
	"github.com/alomerry/go-tools/components/oss/meta"
	innerMinio "github.com/alomerry/go-tools/components/oss/minio"
	"github.com/alomerry/go-tools/components/oss/s3"
)

var (
	ErrUnsupportedClientType = errors.New("unsupported client type")
)

func NewClient(cfg meta.Config) (meta.OSSClient, error) {
	switch cfg.Type {
	case meta.ClientTypeMinio:
		return innerMinio.NewMinioClient(cfg)
	case meta.ClientTypeKodo:
		return innerKodo.NewKodoClient(cfg)
	case meta.ClientTypeR2:
		return s3.NewDefaultCloudflareR2(cfg)
	case meta.ClientTypeS3:
		return s3.NewDefaultCloudflareR2(cfg)
	default:
		return nil, ErrUnsupportedClientType
	}
}
