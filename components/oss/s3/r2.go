package s3

import (
	"context"
	"fmt"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"os"
)

type CloudflareR2 struct {
	client *s3.Client
}

func NewDefaultCloudflareR2() any {
	return NewCloudflareR2(
		env.GetCloudflareAccountId(),
		env.GetCloudflareR2AccountKey(),
		env.GetCloudflareR2AccountSK(),
	)
}

func NewCloudflareR2(accountId, r2Key, r2Secret string) any {
	c := &CloudflareR2{}
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
		log.Fatal(err)
	}

	c.client = s3.NewFromConfig(cfg)

	return c
}

func (c *CloudflareR2) UploadFromLocal(ctx context.Context, bucket, filePath, ossPath string) (any, error) {
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
