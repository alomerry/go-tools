package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env/oss"
	"github.com/alomerry/go-tools/utils/files"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/downloader"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/region"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
)

type client struct {
	cred *credentials.Credentials
	dm   *downloader.DownloadManager
	um   *uploader.UploadManager

	useSSL     bool
	bucketName string
}

type uploadReturnBody struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func NewKodoClient(cfg model.Config) (meta.OSSClient, error) {
	c := &client{
		useSSL:     cfg.SSL,
		bucketName: cfg.BucketName,
	}

	c.cred = credentials.NewCredentials(cfg.AccessKey, cfg.SecretKey)
	return c, nil
}

func (q *client) UploadManager() *uploader.UploadManager {
	if q.um == nil {
		q.um = uploader.NewUploadManager(&uploader.UploadManagerOptions{
			Options: http_client.Options{
				Credentials: q.cred,
				Regions:     region.GetRegionByID("z0", q.useSSL),
			}})
	}

	return q.um
}

// DownloadManager 惰性初始化下载管理器（原实现 q.dm 从未初始化，
// DownloadToFile 必 nil panic）
func (q *client) DownloadManager() *downloader.DownloadManager {
	if q.dm == nil {
		q.dm = downloader.NewDownloadManager(&downloader.DownloadManagerOptions{
			Options: http_client.Options{
				Credentials: q.cred,
				Regions:     region.GetRegionByID("z0", q.useSSL),
			}})
	}

	return q.dm
}

func (q *client) PutObject(ctx context.Context, objectKey string, reader io.Reader, _ int64) error {
	var res uploadReturnBody
	err := q.UploadManager().UploadReader(ctx, reader, &uploader.ObjectOptions{
		BucketName: q.bucketName,
		ObjectName: &objectKey,
		FileName:   files.GetFileName(objectKey),
	}, &res)
	return err
}

func (q *client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (q *client) DownloadToFile(ctx context.Context, objectKey string) (string, error) {
	urlsProvider := downloader.SignURLsProvider(
		downloader.NewStaticDomainBasedURLsProvider([]string{oss.KodoCdnDomain()}),
		downloader.NewCredentialsSigner(q.cred),
		&downloader.SignOptions{TTL: 1 * time.Hour})

	tmpFileName, err := files.CreateTempFile(ctx, cons.EmptyStr, nil)
	if err != nil {
		return "", err
	}

	downloaded, err := q.DownloadManager().DownloadToFile(
		ctx,
		objectKey,
		tmpFileName,
		&downloader.ObjectOptions{
			GenerateOptions:      downloader.GenerateOptions{BucketName: oss.KodoBasicBucket()},
			DownloadURLsProvider: urlsProvider,
		})
	if err != nil {
		return "", err
	}

	if downloaded == 0 {
		// TODO
	}
	return tmpFileName, nil
}

func (q *client) RemoveObject(ctx context.Context, objectKey string) error {
	//TODO implement me
	return errors.New("not implemented")
}

func (q *client) StatObject(ctx context.Context, objectKey string) (meta.ObjectInfo, error) {
	//TODO implement me
	return meta.ObjectInfo{}, errors.New("not implemented")
}

func (q *client) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	//TODO implement me
	return "", errors.New("not implemented")
}

func (q *client) PresignedPutObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	//TODO implement me
	return "", errors.New("not implemented")
}

func (q *client) CreateBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	return errors.New("not implemented")
}

func (q *client) ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]meta.ObjectInfo, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (q *client) RemoveBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	return errors.New("not implemented")
}

// ListBuckets 返回账户下全部空间。七牛没有与 S3 完全对齐的列桶语义：
// BucketsV4 按分页返回空间名与创建时间（Region 传空即查询所有区域），这里
// 循环翻页聚合成 meta.Bucket。
func (q *client) ListBuckets(ctx context.Context) ([]meta.Bucket, error) {
	// storagev2 credentials.Credentials 是 auth.Credentials 的别名，可直接复用
	bm := storage.NewBucketManager(q.cred, nil)

	results := make([]meta.Bucket, 0)
	marker := ""
	for {
		output, err := bm.BucketsV4(&storage.BucketV4Input{Marker: marker})
		if err != nil {
			return nil, err
		}
		for _, bucket := range output.Buckets {
			results = append(results, meta.Bucket{
				Name:         bucket.Name,
				CreationDate: bucket.Ctime,
			})
		}
		if !output.IsTruncated {
			break
		}
		if output.NextMarker == "" {
			// 服务端截断标记异常（truncated 但无续传 marker）：终止防死循环，
			// 返回已聚合的部分结果
			break
		}
		marker = output.NextMarker
	}
	return results, nil
}

// ListObjectsPage 基于 ListFiles 的 marker/limit 分页语义：continuationToken
// 映射 marker，非 recursive 时 delimiter="/" 且 CommonPrefixes 映射为
// DirPrefixes。七牛 limit 上限 1000，越界收敛到 1000；PutTime 为 100ns 粒度
// 时间戳（去低七位即 Unix 秒）。
func (q *client) ListObjectsPage(ctx context.Context, bucketName string, prefix string, continuationToken string, recursive bool, maxKeys int32) (meta.ObjectPage, error) {
	bm := storage.NewBucketManager(q.cred, nil)
	if maxKeys <= 0 || maxKeys > 1000 {
		maxKeys = 1000
	}
	delim := ""
	if !recursive {
		delim = "/"
	}

	// ListFilesWithContext 透传 ctx：调用方取消/超时在下层 HTTP 请求生效
	// （ListFiles 内部写死 context.Background，不感知取消）。
	ret, _, err := bm.ListFilesWithContext(ctx, bucketName,
		storage.ListInputOptionsPrefix(prefix),
		storage.ListInputOptionsDelimiter(delim),
		storage.ListInputOptionsMarker(continuationToken),
		storage.ListInputOptionsLimit(int(maxKeys)))
	if err != nil {
		return meta.ObjectPage{}, err
	}
	entries, commonPrefixes, nextMarker := ret.Items, ret.CommonPrefixes, ret.Marker

	page := meta.ObjectPage{
		Objects:     make([]meta.ObjectInfo, 0, len(entries)),
		DirPrefixes: make([]string, 0, len(commonPrefixes)),
	}
	for _, item := range entries {
		page.Objects = append(page.Objects, meta.ObjectInfo{
			ETag:         item.Hash,
			Key:          item.Key,
			Size:         item.Fsize,
			LastModified: time.Unix(item.PutTime/10000000, 0).UTC(),
			ContentType:  item.MimeType,
		})
	}
	page.DirPrefixes = append(page.DirPrefixes, commonPrefixes...)
	page.NextToken = nextMarker
	return page, nil
}

func (q *client) Bucket(ctx context.Context, bucketName string) meta.OSSClient {
	q.bucketName = bucketName
	return q
}

// --- 桶配置（meta.OSSClient 扩展）：七牛 kodo 无 S3 对齐的桶配置语义，
// 全系返回 ErrBucketConfigUnsupported（调用方降级展示，非失败）。 ---

func (q *client) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	return "", fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketPolicy(ctx context.Context, bucketName string, policyJSON string) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketLifecycle(ctx context.Context, bucketName string) ([]meta.LifecycleRule, error) {
	return nil, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketLifecycle(ctx context.Context, bucketName string, rules []meta.LifecycleRule) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketVersioning(ctx context.Context, bucketName string) (meta.BucketVersioning, error) {
	return meta.BucketVersioning{}, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketEncryption(ctx context.Context, bucketName string) (meta.BucketEncryption, error) {
	return meta.BucketEncryption{}, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketEncryption(ctx context.Context, bucketName string, cfg meta.BucketEncryption) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketCORS(ctx context.Context, bucketName string) ([]meta.CORSRule, error) {
	return nil, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketCORS(ctx context.Context, bucketName string, rules []meta.CORSRule) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketReferer(ctx context.Context, bucketName string) (meta.RefererConfig, error) {
	return meta.RefererConfig{}, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketReferer(ctx context.Context, bucketName string, cfg meta.RefererConfig) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) GetBucketQuota(ctx context.Context, bucketName string) (meta.BucketQuota, error) {
	return meta.BucketQuota{}, fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}

func (q *client) PutBucketQuota(ctx context.Context, bucketName string, quota meta.BucketQuota) error {
	return fmt.Errorf("%w: kodo", meta.ErrBucketConfigUnsupported)
}
