package internal

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(cfg model.Config) (meta.OSSClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, err
	}
	// 仅打印 endpoint：%#v 会展开客户端结构体（含凭据字段），禁止入日志
	log.Printf("MinIO client connected: %s", cfg.Endpoint)
	return &minioClient{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (m *minioClient) DownloadToFile(ctx context.Context, objectKey string) (string, error) {
	//TODO implement me
	return "", errors.New("not implemented")
}

func (m *minioClient) PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error {
	_, err := m.client.PutObject(ctx, m.bucketName, objectKey, reader, objectSize, minio.PutObjectOptions{})
	return err
}

func (m *minioClient) Bucket(ctx context.Context, bucketName string) meta.OSSClient {
	return &minioClient{
		client:     m.client,
		bucketName: bucketName,
	}
}

func (m *minioClient) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, m.bucketName, objectKey, minio.GetObjectOptions{})
}

func (m *minioClient) RemoveObject(ctx context.Context, objectKey string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectKey, minio.RemoveObjectOptions{})
}

func (m *minioClient) StatObject(ctx context.Context, objectKey string) (meta.ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, m.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return meta.ObjectInfo{}, err
	}

	return meta.ObjectInfo{
		ETag:         info.ETag,
		Key:          info.Key,
		Size:         info.Size,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
	}, nil
}

func (m *minioClient) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (m *minioClient) PresignedPutObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, m.bucketName, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (m *minioClient) CreateBucket(ctx context.Context, bucketName string) error {
	return m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (m *minioClient) ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]meta.ObjectInfo, error) {
	objects := m.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	results := make([]meta.ObjectInfo, 0)
	for object := range objects {
		results = append(results, meta.ObjectInfo{
			ETag: object.ETag,
		})
	}

	return results, nil
}

func (m *minioClient) RemoveBucket(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucket(ctx, bucketName)
}

func (m *minioClient) ListBuckets(ctx context.Context) ([]meta.Bucket, error) {
	buckets, err := m.client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]meta.Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		results = append(results, meta.Bucket{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
		})
	}
	return results, nil
}

// ListObjectsPage 暂未实现：minio-go 不透传 continuation-token，StartAfter
// 词法续传近似会让 token 语义退化泄漏进接口，当前无调用方（deck 分页列举只走
// rustfs/r2），需要时再按 S3 原生分页补齐。
func (m *minioClient) ListObjectsPage(ctx context.Context, bucketName string, prefix string, continuationToken string, recursive bool, maxKeys int32) (meta.ObjectPage, error) {
	return meta.ObjectPage{}, errors.New("not implemented")
}
