package oss

import (
  "errors"
  
  "github.com/alomerry/go-tools/components/oss/internal"
  "github.com/alomerry/go-tools/components/oss/meta"
  "github.com/alomerry/go-tools/model"
)

var (
	ErrUnsupportedClientType = errors.New("unsupported client type")
)

func NewClient(cfg model.Config) (meta.OSSClient, error) {
	switch cfg.Type {
  case model.ClientTypeMinio:
    return internal.NewMinioClient(cfg)
  case model.ClientTypeKodo:
    return internal.NewKodoClient(cfg)
  case model.ClientTypeR2:
    return internal.NewDefaultCloudflareR2(cfg)
  case model.ClientTypeS3:
    return internal.NewRustFs(cfg)
  case model.ClientTypeRustFs:
    return internal.NewRustFs(cfg)
	default:
		return nil, ErrUnsupportedClientType
	}
}
