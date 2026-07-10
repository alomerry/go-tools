package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/alomerry/go-tools/components/oss/meta"
  "github.com/alomerry/go-tools/model"
  "github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
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
	oss := newTestRustFs(t)

	rust, ok := oss.(*RustFs)
	assert.True(t, ok)
	fmt.Println(rust.UploadFromLocal(context.TODO(), "/Users/alomerry/workspace/go/go-tools/output/avatar.png", "666.png"))
}

func TestRustFs_RemoveObject(t *testing.T) {
	oss := newTestRustFs(t)

	err := oss.RemoveObject(context.TODO(), "blog")
	assert.Nil(t, err)
}

func TestRustFs_RemoveObject2(t *testing.T) {
	oss := newTestRustFs(t)

	err := oss.RemoveBucket(context.TODO(), cons.OssBucketBlog)
	assert.Nil(t, err)
}
