package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils/files"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type cloudflareR2 struct {
	client *s3.Client
	bucket string
}

func NewDefaultCloudflareR2(cfg meta.Config) (*cloudflareR2, error) {
	if cfg.AccountId == "" || cfg.AccessKey == "" && cfg.SecretKey == "" {
		return nil, errors.New("accountId or accessKey or secretKey is empty")
	}

	return newCloudflareR2(cfg.AccountId, cfg.AccessKey, cfg.SecretKey)
}

func newCloudflareR2(accountId, r2Key, r2Secret string) (*cloudflareR2, error) {
	c := &cloudflareR2{}
	r2Resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
			}, nil
		})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				r2Key,
				r2Secret,
				cons.EmptyStr,
			)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	c.client = s3.NewFromConfig(cfg)

	return c, nil
}

func (c *cloudflareR2) UploadFromLocal(ctx context.Context, bucket, filePath, ossPath string) (any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &ossPath,
		Body:        file,
		ContentType: nil, // TODO
	})
	if err != nil {
		log.Fatal(err)
	}
	return resp, nil
}

func (c *cloudflareR2) PutObject(ctx context.Context, objectKey string, reader io.Reader, objectSize int64) error {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) DownloadToFile(ctx context.Context, objectKey string) (string, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", c.bucket, objectKey)),
	})
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	fileFullPath, err := files.CreateTempFile(ctx, files.GetFileName(objectKey), func(file *os.File) error {
		_, err = io.Copy(file, result.Body)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return fileFullPath, nil
}

func (c *cloudflareR2) RemoveObject(ctx context.Context, objectKey string) error {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) StatObject(ctx context.Context, objectKey string) (meta.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) PresignedGetObject(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) CreateBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) ListObjects(ctx context.Context, bucketName string, prefix string, recursive bool) ([]meta.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) RemoveBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	panic("implement me")
}

func (c *cloudflareR2) Bucket(_ context.Context, bucketName string) meta.OSSClient {
	return &cloudflareR2{
		client: c.client,
		bucket: bucketName,
	}
}
