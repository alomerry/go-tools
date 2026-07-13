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
