package s3

import (
	"context"
	"log"
	"os"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RustFs struct {
	client *s3.Client
}

func NewDefaultRustFs() *RustFs {
	return NewRustFs(
		env.GetRustfsEndpoint(),
		env.GetRustfsAccessKey(),
		env.GetRustfsSecretKey(),
	)
}

func NewRustFs(endpoint, accessKey, secretKey string) *RustFs {
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
		log.Fatal(err)
	}

	// build S3 client
	r.client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return r
}

func (r *RustFs) UploadFromLocal(ctx context.Context, bucket, filePath, ossPath string) (any, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := r.client.PutObject(ctx, &s3.PutObjectInput{
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

func (r *RustFs) RemoveObject(ctx context.Context, bucket, ossKey string) (any, error) {
	resp, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &ossKey,
	})
	if err != nil {
		log.Fatal(err)
	}
	return resp, nil
}

func (r *RustFs) RemoveBucket(ctx context.Context, bucket string) (any, error) {
	resp, err := r.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		log.Fatal(err)
	}
	return resp, nil
}
