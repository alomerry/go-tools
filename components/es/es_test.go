package es

import (
	"os"
	"testing"
)

// es 凭证经环境变量注入，未配置时跳过（避免真实凭证入库/无环境时误报）。
var opts = func() []Option {
	endpoint := os.Getenv("GO_TOOLS_TEST_ES_ENDPOINT")
	apiKey := os.Getenv("GO_TOOLS_TEST_ES_API_KEY")
	if endpoint == "" || apiKey == "" {
		return nil
	}
	return []Option{
		WithEndpoint(endpoint),
		WithAPIKey(apiKey),
	}
}()

func TestNewClient(t *testing.T) {
	if len(opts) == 0 {
		t.Skip("GO_TOOLS_TEST_ES_ENDPOINT / GO_TOOLS_TEST_ES_API_KEY not set")
	}
	client := NewClient(opts...)
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestGetEs(t *testing.T) {
	if len(opts) == 0 {
		t.Skip("GO_TOOLS_TEST_ES_ENDPOINT / GO_TOOLS_TEST_ES_API_KEY not set")
	}
	client := NewClient(opts...)
	if client == nil {
		t.Fatal("client is nil")
	}

	es := client.GetEs()
	if es == nil {
		t.Fatal("es is nil")
	}
	t.Logf("es: %v", es.Indices.Get("test"))
}
