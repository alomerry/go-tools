package kodo

import (
	"context"
	"github.com/alomerry/go-tools/utils/random"
	"os"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/static/env/oss"
)

var (
	cfg = meta.Config{
		Type:      meta.ClientTypeKodo,
		Endpoint:  oss.KodoCdnDomain(),
		AccessKey: oss.KodoAccessKey(),
		SecretKey: oss.KodoSecretKey(),
		SSL:       false,
	}
)

func TestKodoClient_PutObject(t *testing.T) {
	ctx := context.Background()

	objectKey := "pipeline/build/" + time.Now().Format("2006-01-02") + "/build-" + random.String(7) + ".log"
	filePath := "/Users/alomerry/Downloads/tmp/up.txt"

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	client, err := NewKodoClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create kodo client: %v", err)
	}

	err = client.Bucket(ctx, "alomerry").PutObject(ctx, objectKey, file, -1)
	if err != nil {
		t.Fatalf("Failed to put object: %v", err)
	}

	t.Logf("Object %s uploaded successfully", objectKey)
}
