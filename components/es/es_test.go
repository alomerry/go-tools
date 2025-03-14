package es

import (
	"testing"
)

var opts = []Option{
	WithEndpoint("http://localhost:9200"),
	WithAPIKey("elastic:p6]D4}H-004]ArLWVw]>8E-QB"),
}

func TestNewClient(t *testing.T) {
	client := NewClient(opts...)
	if client == nil {
		t.Fatal("client is nil")
	}
}

func TestGetEs(t *testing.T) {
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
