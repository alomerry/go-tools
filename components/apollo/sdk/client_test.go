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
		// resty 仅在 JSON content-type 下解码 SetResult
		w.Header().Set("Content-Type", "application/json")
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

func TestDo_Error_EmptyBodyWrappedAsErrorResponse(t *testing.T) {
	// When Apollo returns an error status with no parseable error fields
	// (message and exception both empty), the fallback branch must still wrap
	// the result as a *ErrorResponse carrying the HTTP status code. Otherwise
	// downstream errors.As would fail and a 403 would be surfaced as a 500.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")
	var result struct{}
	resp, err := client.Do(context.Background(), http.MethodPost, "/test", nil, &result)
	assert.Error(t, err)
	assert.NotNil(t, resp)

	var apiErr *ErrorResponse
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.Status)
	// The raw body must be preserved so the error is never empty.
	assert.Contains(t, err.Error(), "403")
	assert.NotEmpty(t, apiErr.Message)
}

func TestDo_Error_PreservesFullDetails(t *testing.T) {
	// Apollo returns a 403 with full details (message, exception, timestamp)
	// when a token lacks publish permission. Ensure all fields are preserved.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":403,"message":"access is deny","exception":"com.ctrip.framework.apollo.openapi.service.exception.OpenApiException: token has no publish permission","timestamp":"2026-07-24T10:00:00.000+0800"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token")
	var result struct{}
	resp, err := client.Do(context.Background(), http.MethodPost, "/test", nil, &result)
	assert.Error(t, err)
	assert.NotNil(t, resp)

	var apiErr *ErrorResponse
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.Status)
	assert.Equal(t, "access is deny", apiErr.Message)
	assert.Contains(t, apiErr.Exception, "no publish permission")
	assert.Equal(t, "2026-07-24T10:00:00.000+0800", apiErr.Timestamp)
	// Error string must surface every field so callers see the root cause.
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "access is deny")
	assert.Contains(t, err.Error(), "no publish permission")
	assert.Contains(t, err.Error(), "2026-07-24T10:00:00.000+0800")
}
