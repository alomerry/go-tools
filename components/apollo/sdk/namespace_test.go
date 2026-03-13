package sdk

import (
  "context"
  "testing"
  
  "github.com/alomerry/go-tools/static/env"
  "github.com/stretchr/testify/assert"
)

type fullInfo struct {
  appId     string
  cluster   string
  namespace string
  env       string
}

func (f fullInfo) GetAppId() string {
  return f.appId
}

func (f fullInfo) GetCluster() string {
  return f.cluster
}

func (f fullInfo) GetEnv() string {
  return f.env
}

func (f fullInfo) GetNamespace() string {
  return f.namespace
}

func TestGetNamespace(t *testing.T) {
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
  res, err:= client.Namespaces.Get(ctx, req)
  assert.NoError(t, err)
  assert.NotNil(t, res)
}
