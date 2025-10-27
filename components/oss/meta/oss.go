package meta

import (
	"context"
	"io"
	"time"
)

// ClientType 存储类型枚举
type ClientType string

const (
	ClientTypeMinio  ClientType = "minio"
	ClientTypeR2     ClientType = "r2"
	ClientTypeKodo   ClientType = "kodo"
	ClientTypeRustFs ClientType = "rust-fs"
	ClientTypeS3     ClientType = "s3"
)

// Config 存储配置
type Config struct {
	Type       ClientType // 客户端类型
	Endpoint   string     // 服务端点
	AccessKey  string     // 访问密钥
	SecretKey  string     // 秘密密钥
	Region     string     // 区域（S3 需要）
	AccountId  string     // 账户 ID（R2 需要）
	SSL        bool       // 是否使用 SSL
	BucketName string     // 默认存储桶名称
}

// OSSClient 抽象接口
type OSSClient interface {
	PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadToFile(ctx context.Context, objectKey string) (string, error)
	RemoveObject(ctx context.Context, objectKey string) error
	StatObject(ctx context.Context, objectKey string) (ObjectInfo, error)
	PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	CreateBucket(ctx context.Context, bucketName string) error
	ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]ObjectInfo, error)
	RemoveBucket(ctx context.Context, bucketName string) error
	Bucket(ctx context.Context, bucketName string) OSSClient
}

// ObjectInfo 对象元数据
type ObjectInfo struct {
	ETag         string
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}
