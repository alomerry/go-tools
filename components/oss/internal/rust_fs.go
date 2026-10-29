package internal

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/alomerry/go-tools/components/ext"
	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/utils/files"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RustFs struct {
	client *s3.Client
	bucket string
}

func NewDefaultRustFs() (meta.OSSClient, error) {
	return newRustFs(
		env.GetRustfsEndpoint(),
		env.GetRustfsAccessKey(),
		env.GetRustfsSecretKey(),
	)
}

func NewRustFs(cfg model.Config) (meta.OSSClient, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		// return nil, errors.New("endpoint or accessKey or secretKey is empty")
		cfg = *ext.Apollo().RustFs()
	}

	return newRustFs(cfg.Endpoint, cfg.AccessKey, cfg.SecretKey)
}

func newRustFs(endpoint, accessKey, secretKey string) (*RustFs, error) {
	r := &RustFs{}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				cons.EmptyStr,
			)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	// build S3 client
	r.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return r, nil
}

// UploadFromLocal uploads a local file to the bound bucket. It is a convenience
// helper on top of PutObject and is not part of the meta.OSSClient contract.
func (r *RustFs) UploadFromLocal(ctx context.Context, filePath, ossPath string) (any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	resp, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &ossPath,
		Body:   file,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *RustFs) PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error {
	input := &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
		Body:   reader,
	}
	if objectSize > 0 {
		input.ContentLength = aws.Int64(objectSize)
	}
	_, err := r.client.PutObject(ctx, input)
	return err
}

func (r *RustFs) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (r *RustFs) DownloadToFile(ctx context.Context, objectKey string) (string, error) {
	resp, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return files.CreateTempFile(ctx, files.GetFileName(objectKey), func(file *os.File) error {
		_, err := io.Copy(file, resp.Body)
		return err
	})
}

func (r *RustFs) RemoveObject(ctx context.Context, objectKey string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	})
	return err
}

func (r *RustFs) StatObject(ctx context.Context, objectKey string) (meta.ObjectInfo, error) {
	resp, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return meta.ObjectInfo{}, err
	}

	info := meta.ObjectInfo{
		Key:          objectKey,
		Size:         aws.ToInt64(resp.ContentLength),
		LastModified: aws.ToTime(resp.LastModified),
	}
	if resp.ETag != nil {
		info.ETag = *resp.ETag
	}
	if resp.ContentType != nil {
		info.ContentType = *resp.ContentType
	}
	return info, nil
}

func (r *RustFs) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(r.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *RustFs) PresignedPutObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(r.client)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &r.bucket,
		Key:    &objectKey,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *RustFs) CreateBucket(ctx context.Context, bucketName string) error {
	_, err := r.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &bucketName})
	return err
}

func (r *RustFs) ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]meta.ObjectInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}
	if prefix != "" {
		input.Prefix = &prefix
	}
	// When not recursive, use "/" as the delimiter so only the current
	// "directory" level is returned (CommonPrefixes holds the sub-prefixes).
	if !recursive {
		delim := "/"
		input.Delimiter = &delim
	}

	paginator := s3.NewListObjectsV2Paginator(r.client, input)
	results := make([]meta.ObjectInfo, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			info := meta.ObjectInfo{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			}
			if obj.ETag != nil {
				info.ETag = *obj.ETag
			}
			results = append(results, info)
		}
	}
	return results, nil
}

func (r *RustFs) RemoveBucket(ctx context.Context, bucketName string) error {
	_, err := r.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucketName,
	})
	return err
}

func (r *RustFs) ListBuckets(ctx context.Context) ([]meta.Bucket, error) {
	paginator := s3.NewListBucketsPaginator(r.client, &s3.ListBucketsInput{})
	results := make([]meta.Bucket, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, bucket := range page.Buckets {
			results = append(results, meta.Bucket{
				Name:         aws.ToString(bucket.Name),
				CreationDate: aws.ToTime(bucket.CreationDate),
			})
		}
	}
	return results, nil
}

func (r *RustFs) ListObjectsPage(ctx context.Context, bucketName string, prefix string, continuationToken string, recursive bool, maxKeys int32) (meta.ObjectPage, error) {
	// 与 R2 共用 ListObjectsV2 单页语义（见 listObjectsV2Page）
	return listObjectsV2Page(ctx, r.client, bucketName, prefix, continuationToken, recursive, maxKeys)
}

// listObjectsV2Page 执行一次 ListObjectsV2 单页调用：continuationToken/
// maxKeys 原样透传，非 recursive 时 delimiter="/"，CommonPrefixes 映射为
// DirPrefixes，NextContinuationToken 为空表示没有更多页。RustFs 与 R2 共用。
func listObjectsV2Page(ctx context.Context, client *s3.Client, bucketName string, prefix string, continuationToken string, recursive bool, maxKeys int32) (meta.ObjectPage, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}
	if prefix != "" {
		input.Prefix = &prefix
	}
	if continuationToken != "" {
		input.ContinuationToken = &continuationToken
	}
	if maxKeys > 0 {
		input.MaxKeys = &maxKeys
	}
	if !recursive {
		delim := "/"
		input.Delimiter = &delim
	}

	output, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		return meta.ObjectPage{}, err
	}

	page := meta.ObjectPage{
		Objects:     make([]meta.ObjectInfo, 0, len(output.Contents)),
		DirPrefixes: make([]string, 0, len(output.CommonPrefixes)),
	}
	for _, obj := range output.Contents {
		info := meta.ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
		}
		if obj.ETag != nil {
			info.ETag = *obj.ETag
		}
		page.Objects = append(page.Objects, info)
	}
	for _, cp := range output.CommonPrefixes {
		page.DirPrefixes = append(page.DirPrefixes, aws.ToString(cp.Prefix))
	}
	page.NextToken = aws.ToString(output.NextContinuationToken)
	return page, nil
}

// Bucket returns a new RustFs bound to the given bucket, sharing the same S3
// client. The returned client operates only on that bucket for the
// object-level methods (PutObject/GetObject/...); bucket-level methods
// (CreateBucket/RemoveBucket/ListObjects) still take an explicit bucketName.
func (r *RustFs) Bucket(_ context.Context, bucketName string) meta.OSSClient {
	return &RustFs{
		client: r.client,
		bucket: bucketName,
	}
}

// --- 桶配置（meta.OSSClient 扩展）：全部委托 internal/bucket_config.go 的
// S3 共享实现，双实例（rustfs/R2）同语义；R2 不支持项由共享层归一为
// meta.ErrBucketConfigUnsupported。 ---

func (r *RustFs) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	return s3GetBucketPolicy(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketPolicy(ctx context.Context, bucketName string, policyJSON string) error {
	return s3PutBucketPolicy(ctx, r.client, bucketName, policyJSON)
}

func (r *RustFs) GetBucketLifecycle(ctx context.Context, bucketName string) ([]meta.LifecycleRule, error) {
	return s3GetBucketLifecycle(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketLifecycle(ctx context.Context, bucketName string, rules []meta.LifecycleRule) error {
	return s3PutBucketLifecycle(ctx, r.client, bucketName, rules)
}

func (r *RustFs) GetBucketVersioning(ctx context.Context, bucketName string) (meta.BucketVersioning, error) {
	return s3GetBucketVersioning(ctx, r.client, bucketName)
}

func (r *RustFs) SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error {
	return s3SetBucketVersioning(ctx, r.client, bucketName, enabled)
}

func (r *RustFs) GetBucketEncryption(ctx context.Context, bucketName string) (meta.BucketEncryption, error) {
	return s3GetBucketEncryption(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketEncryption(ctx context.Context, bucketName string, cfg meta.BucketEncryption) error {
	return s3PutBucketEncryption(ctx, r.client, bucketName, cfg)
}

func (r *RustFs) GetBucketCORS(ctx context.Context, bucketName string) ([]meta.CORSRule, error) {
	return s3GetBucketCORS(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketCORS(ctx context.Context, bucketName string, rules []meta.CORSRule) error {
	return s3PutBucketCORS(ctx, r.client, bucketName, rules)
}

func (r *RustFs) GetBucketReferer(ctx context.Context, bucketName string) (meta.RefererConfig, error) {
	return s3GetBucketReferer(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketReferer(ctx context.Context, bucketName string, cfg meta.RefererConfig) error {
	return s3PutBucketReferer(ctx, r.client, bucketName, cfg)
}

func (r *RustFs) GetBucketQuota(ctx context.Context, bucketName string) (meta.BucketQuota, error) {
	return s3GetBucketQuota(ctx, r.client, bucketName)
}

func (r *RustFs) PutBucketQuota(ctx context.Context, bucketName string, quota meta.BucketQuota) error {
	return s3PutBucketQuota(ctx, r.client, bucketName, quota)
}
