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
