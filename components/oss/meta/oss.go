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
