package sdk

import (
  "context"
  "testing"
  
  "github.com/alomerry/go-tools/static/cons/apollo"
  "github.com/alomerry/go-tools/static/env"
  "github.com/stretchr/testify/assert"
)

func TestGetItem(t *testing.T) {
  var (
    ctx = context.TODO()
    req = fullInfo{
      appId:     "homelab",
      cluster:   "default",
      namespace: "application",
      env:       "dev",
    }
  )
  client := NewClient("https://apollo.alomerry.cn", env.ApolloOpenapiToken())
  assert.NotNil(t, client)
  res, err := client.Items.Get(ctx, req, "backend.openapi.meta")
  assert.NoError(t, err)
  assert.NotNil(t, res)
  
  list, err := client.Items.List(ctx, req, 1, 10)
  assert.NoError(t, err)
  assert.NotNil(t, list)
}

func TestUpdateItem(t *testing.T) {
  var (
    ctx = context.TODO()
    req = fullInfo{
      appId:     "homelab",
      cluster:   "default",
      namespace: "application",
      env:       "dev",
    }
  )
  client := NewClient("https://apollo.alomerry.cn", env.ApolloOpenapiToken())
  assert.NotNil(t, client)
  res, err := client.Items.Update(ctx, req, &Item{
    Key:                      "backend.openapi.meta",
    Value:                    "{\"port\": 8091}",
    Type:                     apollo.KeyTypeJson,
    Comment:                  "meta 配置",
    DataChangeCreatedBy:      "homelab",
    DataChangeLastModifiedBy: "homelab",
  })
  assert.NoError(t, err)
  assert.NotNil(t, res)
}