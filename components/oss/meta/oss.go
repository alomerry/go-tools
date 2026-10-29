package meta

import (
	"context"
	"io"
	"time"
)

type OSSClient interface {
	PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadToFile(ctx context.Context, objectKey string) (string, error)
	RemoveObject(ctx context.Context, objectKey string) error
	StatObject(ctx context.Context, objectKey string) (ObjectInfo, error)
	PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	PresignedPutObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	CreateBucket(ctx context.Context, bucketName string) error
	ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]ObjectInfo, error)
	ListObjectsPage(ctx context.Context, bucketName string, prefix string, continuationToken string, recursive bool, maxKeys int32) (ObjectPage, error)
	ListBuckets(ctx context.Context) ([]Bucket, error)
	RemoveBucket(ctx context.Context, bucketName string) error
	Bucket(ctx context.Context, bucketName string) OSSClient

	// 桶配置（bucket_config.go）：BucketPolicy/Lifecycle/Versioning/
	// Encryption/CORS 走 S3 原生 API；Referer/Quota 为应用层语义（Bucket
	// Tagging 透存，见 RefererConfig/BucketQuota 注释）。不支持的后端返回
	// ErrBucketConfigUnsupported（调用方降级展示，非失败）。
	// 清空语义：Put 空值即删（policy 空串 / lifecycle 空集 / encryption
	// algorithm 空 / cors 空集走 S3 Delete），不设独立 Delete 方法。
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)
	PutBucketPolicy(ctx context.Context, bucketName string, policyJSON string) error
	GetBucketLifecycle(ctx context.Context, bucketName string) ([]LifecycleRule, error)
	PutBucketLifecycle(ctx context.Context, bucketName string, rules []LifecycleRule) error
	GetBucketVersioning(ctx context.Context, bucketName string) (BucketVersioning, error)
	SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error
	GetBucketEncryption(ctx context.Context, bucketName string) (BucketEncryption, error)
	PutBucketEncryption(ctx context.Context, bucketName string, cfg BucketEncryption) error
	GetBucketCORS(ctx context.Context, bucketName string) ([]CORSRule, error)
	PutBucketCORS(ctx context.Context, bucketName string, rules []CORSRule) error
	GetBucketReferer(ctx context.Context, bucketName string) (RefererConfig, error)
	PutBucketReferer(ctx context.Context, bucketName string, cfg RefererConfig) error
	GetBucketQuota(ctx context.Context, bucketName string) (BucketQuota, error)
	PutBucketQuota(ctx context.Context, bucketName string, quota BucketQuota) error
}

// Bucket 存储桶元数据
type Bucket struct {
	Name         string
	CreationDate time.Time
}

// ObjectPage 单页对象列举结果（ListObjectsV2 语义）：Objects 为当前页对象，
// DirPrefixes 为 CommonPrefixes（仅非 recursive 时返回），NextToken 为空表示
// 没有更多页。
type ObjectPage struct {
	Objects     []ObjectInfo
	DirPrefixes []string
	NextToken   string
}

// ObjectInfo 对象元数据
type ObjectInfo struct {
	ETag         string
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}
