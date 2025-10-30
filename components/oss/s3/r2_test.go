package s3

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestNewCloudflareR2_DownloadToFileByCloudflareR2(t *testing.T) {
	oss, err := newCloudflareR2(
		env.GetCloudflareAccountId(),
		env.GetCloudflareR2AccountKey(),
		env.GetCloudflareR2AccountSK(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, oss)
	// oss.UploadFromLocal(context.TODO(), cons.OssBucketCdn, "/Users/alomerry/workspace/go-tools/output/avatar.jpg", "blog/666.jpg")
}

func TestCloudflareR2_DownloadToFileByCloudflareR2(t *testing.T) {
	oss, err := newCloudflareR2(
		env.GetCloudflareAccountId(),
		env.GetCloudflareR2AccountKey(),
		env.GetCloudflareR2AccountSK(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, oss)

	ctx := context.TODO()

	filePath, err := oss.Bucket(ctx, cons.OssBucketCdn).DownloadToFile(ctx, "backup/blog/markdowns.tar.gz")
	assert.Nil(t, err)
	assert.Greater(t, len(filePath), 0)
	t.Log(filePath)
}
