package redis

import (
	"context"
	"testing"
  
  "github.com/alomerry/go-tools/static/cons"
  "github.com/stretchr/testify/assert"
)

var (
	url = ""
)

const (
  testGen1 = "1"
)

func TestGenRedisKey(t *testing.T) {
	client := NewRedisClient(url)
	res := client.Get(context.TODO(), "whoami")
	assert.NotNil(t, res)
	assert.Equal(t, "homelab", res.Val())
}

func TestKeyGen(t *testing.T) {
  assert.NotEmpty(t, KeyGen().GenKey("test1"))
  assert.NotEmpty(t, KeyGen().GenKey(testGen1))
  assert.NotEmpty(t, KeyGen().GenKey(cons.RedisKeyCategoryDefault))
}