package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
  
  "github.com/alomerry/go-tools/static/env"
  "github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
  client := NewClient(env.ApolloHost(), env.ApolloOpenapiToken())
	assert.NotNil(t, client)
  assert.Equal(t, env.ApolloHost(), client.client.BaseURL)
  assert.Equal(t, env.ApolloOpenapiToken(), client.client.Header.Get("Authorization"))
  
}

func TestDo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")
	var result struct {
		Name string `json:"name"`
	}
	resp, err := client.Do(context.Background(), http.MethodGet, "/test", nil, &result)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test", result.Name)
}

func TestDo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":400,"message":"bad request"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")
	var result struct{}
	resp, err := client.Do(context.Background(), http.MethodGet, "/test", nil, &result)
	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "bad request")
}
