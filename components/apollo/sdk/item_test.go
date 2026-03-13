package sdk

import (
  "context"
  "testing"
  
  "github.com/alomerry/go-tools/static/env"
  "github.com/stretchr/testify/assert"
)

func TestGetItem(t *testing.T) {
  var (
    ctx = context.TODO()
  )
  client := NewClient("https://apollo.alomerry.cn", env.ApolloOpenapiToken())
  assert.NotNil(t, client)
  res, err:= client.Items.Get(ctx, "dev", "homelab", "default", "application", "backend.openapi.meta")
  assert.NoError(t, err)
  assert.NotNil(t, res)
  
  list, err := client.Items.List(ctx, "dev", "homelab", "default", "application", 1,10)
  assert.NoError(t, err)
  assert.NotNil(t, list)
}