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
    // S3 走 minio-go（标准 S3 兼容协议客户端）；原路由到 NewRustFs
    // 会读 RustFS 专属配置导致 S3 用户连错后端
    return internal.NewMinioClient(cfg)
  case model.ClientTypeRustFs:
    return internal.NewRustFs(cfg)
	default:
		return nil, ErrUnsupportedClientType
	}
}
