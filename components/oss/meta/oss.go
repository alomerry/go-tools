package meta

import (
	"context"
	"io"
	"time"
)

// ClientType 存储类型枚举
type ClientType string

const (
	ClientTypeMinio ClientType = "minio"
	ClientKodo      ClientType = "kodo"
)

// Config 存储配置
type Config struct {
	Type       ClientType // 客户端类型
	Endpoint   string     // 服务端点
	AccessKey  string     // 访问密钥
	SecretKey  string     // 秘密密钥
	Region     string     // 区域（S3需要）
	SSL        bool       // 是否使用SSL
	BucketName string     // 默认存储桶名称
}

// OSSClient 抽象接口
type OSSClient interface {
	PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadToFile(ctx context.Context, objectKey string) (fileName string, err error)
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
