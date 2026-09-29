package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
)

func newTestRustFs(t *testing.T) meta.OSSClient {
	c, err := NewRustFs(model.Config{
		Type:      model.ClientTypeRustFs,
		Endpoint:  env.GetRustfsEndpoint(),
		AccessKey: env.GetRustfsAccessKey(),
		SecretKey: env.GetRustfsSecretKey(),
	})
	assert.Nil(t, err)
	return c.Bucket(context.TODO(), cons.OssBucketBlog)
}

func TestUploadFromLocalByRustFs(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	oss := newTestRustFs(t)

	rust, ok := oss.(*RustFs)
	assert.True(t, ok)
	fmt.Println(rust.UploadFromLocal(context.TODO(), "/Users/alomerry/workspace/go/go-tools/output/avatar.png", "666.png"))
}

func TestRustFs_RemoveObject(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	oss := newTestRustFs(t)

	err := oss.RemoveObject(context.TODO(), "blog")
	assert.Nil(t, err)
}

func TestRustFs_RemoveObject2(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	oss := newTestRustFs(t)

	err := oss.RemoveBucket(context.TODO(), cons.OssBucketBlog)
	assert.Nil(t, err)
}

func TestRustFs_ListBuckets(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	oss := newTestRustFs(t)

	buckets, err := oss.ListBuckets(context.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, buckets)
	for _, bucket := range buckets {
		t.Logf("Bucket %s created at %s", bucket.Name, bucket.CreationDate)
	}
}

func TestRustFs_ListObjectsPage(t *testing.T) {
	if !test.IntegrationEnabled() {
		t.Skip("integration test; set GO_TOOLS_TEST_INTEGRATION=1 to enable")
	}
	oss := newTestRustFs(t)

	// 目录浏览（delimiter）语义：DirPrefixes 承载 CommonPrefixes
	page, err := oss.ListObjectsPage(context.TODO(), cons.OssBucketBlog, "", "", false, 10)
	assert.Nil(t, err)
	for _, obj := range page.Objects {
		t.Logf("object key=%s size=%d", obj.Key, obj.Size)
	}
	for _, dir := range page.DirPrefixes {
		t.Logf("dir prefix=%s", dir)
	}
	t.Logf("next token=%q", page.NextToken)
}
