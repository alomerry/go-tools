package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	url = ""
)

func TestGenRedisKey(t *testing.T) {
	client := NewRedisClient(url)
	res := client.Get(context.TODO(), "whoami")
	assert.NotNil(t, res)
	assert.Equal(t, "homelab", res.Val())
}
