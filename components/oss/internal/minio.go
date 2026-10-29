package internal

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioClient struct {
	client     *minio.Client
	bucketName string

	endpoint string
	access   string
	secret   string
	useSSL   bool

	// s3holder 是桶配置（Policy/Lifecycle/…）走的 AWS S3 SDK 客户端的
	// 惰性持有者：minio-go 的 removeBucketPolicy/removeBucketLifecycle/
	// removeBucketCors 均为私有方法，删除语义无法经公开 API 表达；S3 桶配置
	// API 全族经由本客户端，与 rustfs/R2 共用 internal/bucket_config.go 实现。
	// 必须经指针共享（Bucket() 派生新 struct 时同指针透传，否则派生 clients
	// 各自重复构造且绕过初始化互斥）。
	s3holder *s3ConfigHolder
}

// s3ConfigHolder 持有惰性构造的 AWS S3 SDK 客户端（mutex 守卫，无网络 I/O）。
// 失败不缓存：LoadDefaultConfig 失败时不置 inited，下次调用重试（sync.Once
// 会永久缓存失败，首个暂态错误即判死刑——见 oss.go 后端侧同口径注释）。
type s3ConfigHolder struct {
	mu       sync.Mutex
	endpoint string
	access   string
	secret   string
	useSSL   bool
	client   *s3.Client
	inited   bool
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
		endpoint:   cfg.Endpoint,
		access:     cfg.AccessKey,
		secret:     cfg.SecretKey,
		useSSL:     cfg.SSL,
		s3holder: &s3ConfigHolder{
			endpoint: cfg.Endpoint,
			access:   cfg.AccessKey,
			secret:   cfg.SecretKey,
			useSSL:   cfg.SSL,
		},
	}, nil
}

// s3ConfigClient 返回桶配置用的 AWS S3 SDK 客户端（惰性构造、无网络 I/O）。
// endpoint 形态兼容 "host:port" 与 "http(s)://host[:port]" 两种（minio.New
// 同口径）；缺 scheme 时按 SSL 补齐。构造失败不缓存，下次调用重试。
func (m *minioClient) s3ConfigClient(ctx context.Context) (*s3.Client, error) {
	h := m.s3holder
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.inited {
		return h.client, nil
	}
	endpoint := h.endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		scheme := "http://"
		if h.useSSL {
			scheme = "https://"
		}
		endpoint = scheme + endpoint
	}
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint}, nil
		})
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(
			awscreds.NewStaticCredentialsProvider(h.access, h.secret, "")),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, err
	}
	h.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	h.inited = true
	return h.client, nil
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
		endpoint:   m.endpoint,
		access:     m.access,
		secret:     m.secret,
		useSSL:     m.useSSL,
		// s3 配置客户端经指针共享：派生体复用同一惰性构造结果（含初始化
		// 互斥），不重复构造、不绕锁（见 s3ConfigHolder 注释）。
		s3holder: m.s3holder,
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

// --- 桶配置（meta.OSSClient 扩展）：经 AWS S3 SDK 客户端走 S3 原生 API，
// 与 rustfs/R2 共用 internal/bucket_config.go 实现（含清空语义：Put 空值
// 即删，不设独立 Delete）；kodo 外的
// S3 兼容后端（MinIO/通用 S3）同语义。Tagging 不可用的后端由共享层归一为
// meta.ErrBucketConfigUnsupported。 ---

// s3do 取 S3 配置客户端并执行 fn（Get 类：返回值 + error）。
func s3do[T any](m *minioClient, ctx context.Context, fn func(*s3.Client) (T, error)) (T, error) {
	cli, err := m.s3ConfigClient(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	return fn(cli)
}

// s3put 取 S3 配置客户端并执行写 fn（Put/Set 类：仅 error）。
func (m *minioClient) s3put(ctx context.Context, fn func(*s3.Client) error) error {
	cli, err := m.s3ConfigClient(ctx)
	if err != nil {
		return err
	}
	return fn(cli)
}

func (m *minioClient) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	return s3do(m, ctx, func(cli *s3.Client) (string, error) {
		return s3GetBucketPolicy(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketPolicy(ctx context.Context, bucketName string, policyJSON string) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketPolicy(ctx, cli, bucketName, policyJSON)
	})
}

func (m *minioClient) GetBucketLifecycle(ctx context.Context, bucketName string) ([]meta.LifecycleRule, error) {
	return s3do(m, ctx, func(cli *s3.Client) ([]meta.LifecycleRule, error) {
		return s3GetBucketLifecycle(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketLifecycle(ctx context.Context, bucketName string, rules []meta.LifecycleRule) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketLifecycle(ctx, cli, bucketName, rules)
	})
}

func (m *minioClient) GetBucketVersioning(ctx context.Context, bucketName string) (meta.BucketVersioning, error) {
	return s3do(m, ctx, func(cli *s3.Client) (meta.BucketVersioning, error) {
		return s3GetBucketVersioning(ctx, cli, bucketName)
	})
}

func (m *minioClient) SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3SetBucketVersioning(ctx, cli, bucketName, enabled)
	})
}

func (m *minioClient) GetBucketEncryption(ctx context.Context, bucketName string) (meta.BucketEncryption, error) {
	return s3do(m, ctx, func(cli *s3.Client) (meta.BucketEncryption, error) {
		return s3GetBucketEncryption(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketEncryption(ctx context.Context, bucketName string, cfg meta.BucketEncryption) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketEncryption(ctx, cli, bucketName, cfg)
	})
}

func (m *minioClient) GetBucketCORS(ctx context.Context, bucketName string) ([]meta.CORSRule, error) {
	return s3do(m, ctx, func(cli *s3.Client) ([]meta.CORSRule, error) {
		return s3GetBucketCORS(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketCORS(ctx context.Context, bucketName string, rules []meta.CORSRule) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketCORS(ctx, cli, bucketName, rules)
	})
}

func (m *minioClient) GetBucketReferer(ctx context.Context, bucketName string) (meta.RefererConfig, error) {
	return s3do(m, ctx, func(cli *s3.Client) (meta.RefererConfig, error) {
		return s3GetBucketReferer(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketReferer(ctx context.Context, bucketName string, cfg meta.RefererConfig) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketReferer(ctx, cli, bucketName, cfg)
	})
}

func (m *minioClient) GetBucketQuota(ctx context.Context, bucketName string) (meta.BucketQuota, error) {
	return s3do(m, ctx, func(cli *s3.Client) (meta.BucketQuota, error) {
		return s3GetBucketQuota(ctx, cli, bucketName)
	})
}

func (m *minioClient) PutBucketQuota(ctx context.Context, bucketName string, quota meta.BucketQuota) error {
	return m.s3put(ctx, func(cli *s3.Client) error {
		return s3PutBucketQuota(ctx, cli, bucketName, quota)
	})
}
