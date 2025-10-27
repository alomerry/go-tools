package s3

import (
	"context"
	"fmt"
	"testing"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestUploadFromLocalByRustFs(t *testing.T) {
	oss := NewRustFs(
		env.GetRustfsEndpoint(),
		env.GetRustfsAccessKey(),
		env.GetRustfsSecretKey(),
	)

	fmt.Println(oss.UploadFromLocal(context.TODO(), cons.OssBucketBlog, "/Users/alomerry/workspace/go/go-tools/output/avatar.png", "666.png"))
}

func TestRustFs_RemoveObject(t *testing.T) {
	oss := NewRustFs(
		env.GetRustfsEndpoint(),
		env.GetRustfsAccessKey(),
		env.GetRustfsSecretKey(),
	)

	result, err := oss.RemoveObject(context.TODO(), cons.OssBucketBlog, "blog")
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestRustFs_RemoveObject2(t *testing.T) {
	oss := NewRustFs(
		env.GetRustfsEndpoint(),
		env.GetRustfsAccessKey(),
		env.GetRustfsSecretKey(),
	)

	result, err := oss.RemoveBucket(context.TODO(), cons.OssBucketBlog)
	assert.Nil(t, err)
	fmt.Println(result)
}
