package minio

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/static/env/oss"
)

var (
	cfg = meta.Config{
		Type:      meta.ClientTypeMinio,
		Endpoint:  oss.MinioEndpoint(),
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
	}
)

func TestMinioClient_PutObject(t *testing.T) {
	cfg.SSL = false

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	objectKey := "homelab/test.txt"
	objectContent := "Hello, Minio!"

	err = client.Bucket(context.Background(), "homelab").PutObject(context.Background(), objectKey, strings.NewReader(objectContent), int64(len(objectContent)))
	if err != nil {
		t.Fatalf("Failed to put object: %v", err)
	}

	t.Logf("Object %s uploaded successfully", objectKey)
}

func TestMinioClient_GetObject(t *testing.T) {
	cfg.SSL = false
	cfg.BucketName = "homelab"

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	objectKey := "homelab/test.txt"
	reader, err := client.GetObject(context.Background(), objectKey)
	if err != nil {
		t.Fatalf("Failed to get object: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read object content: %v", err)
	}

	t.Logf("Object %s content: %s", objectKey, string(content))
}

func TestMinioClient_RemoveObject(t *testing.T) {
	cfg.SSL = false
	cfg.BucketName = "homelab"

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	objectKey := "homelab/test.txt"

	err = client.RemoveObject(context.Background(), objectKey)
	if err != nil {
		t.Fatalf("Failed to remove object: %v", err)
	}

	t.Logf("Object %s removed successfully", objectKey)
}

func TestMinioClient_StatObject(t *testing.T) {
	cfg.SSL = false
	cfg.BucketName = "homelab"

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	objectKey := "homelab/test.txt"

	info, err := client.StatObject(context.Background(), objectKey)
	if err != nil {
		t.Fatalf("Failed to stat object: %v", err)
	}

	t.Logf("Object %s info: %+v", objectKey, info)
}

func TestMinioClient_PresignedGetObject(t *testing.T) {
	cfg := meta.Config{
		Endpoint:   "localhost:9000",
		AccessKey:  "minioadmin",
		SecretKey:  "minioadmin",
		BucketName: "test-bucket",
	}

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	objectKey := "test-object"

	presignedURL, err := client.PresignedGetObject(context.Background(), objectKey, 3600*time.Second)
	if err != nil {
		t.Fatalf("Failed to get presigned URL: %v", err)
	}

	t.Logf("Presigned URL for object %s: %s", objectKey, presignedURL)
}

func TestMinioClient_CreateBucket(t *testing.T) {
	cfg.SSL = false

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	err = client.CreateBucket(context.Background(), "test-bucket")
	if err != nil {
		t.Fatalf("Failed to create bucket: %v", err)
	}

	t.Logf("Bucket %s created successfully", "test-bucket")
}

func TestMinioClient_ListObjects(t *testing.T) {
	// cfg := meta.Config{
	// 	Endpoint:   "localhost:9000",
	// 	AccessKey:  "minioadmin",
	// 	SecretKey:  "minioadmin",
	// 	BucketName: "test-bucket",
	// }

	// client, err := NewMinioClient(cfg)
	// if err != nil {
	// 	t.Fatalf("Failed to create minio client: %v", err)
	// }

	// objects, err := client.ListObjects(context.Background(), "", true)
	// if err != nil {
	// 	t.Fatalf("Failed to list objects: %v", err)
	// }

	// t.Logf("Objects in bucket %s: %v", cfg.BucketName, objects)
}

func TestMinioClient_CopyObject(t *testing.T) {
	// cfg := meta.Config{
	// 	Endpoint:   "localhost:9000",
	// 	AccessKey:  "minioadmin",
	// 	SecretKey:  "minioadmin",
	// 	BucketName: "test-bucket",
	// }

	// client, err := NewMinioClient(cfg)
	// if err != nil {
	// 	t.Fatalf("Failed to create minio client: %v", err)
	// }

	// sourceKey := "test-object"
	// destinationKey := "test-object-copy"

	// err = client.CopyObject(context.Background(), sourceKey, destinationKey)
	// if err != nil {
	// 	t.Fatalf("Failed to copy object: %v", err)
	// }

	// t.Logf("Object %s copied to %s successfully", sourceKey, destinationKey)
}

func TestMinioClient_RemoveBucket(t *testing.T) {
	cfg.SSL = false

	client, err := NewMinioClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	err = client.RemoveBucket(context.Background(), "test-bucket")
	if err != nil {
		t.Fatalf("Failed to remove bucket: %v", err)
	}

	t.Logf("Bucket %s removed successfully", "test-bucket")
}
