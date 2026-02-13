package sdk

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/es"
	"github.com/stretchr/testify/assert"
)

func TestSearchLogs(t *testing.T) {
	// Skip integration test by default
	t.Skip("Skipping integration test requiring ES connection")

	client := es.NewClient(es.WithEndpoint("http://localhost:9200"))
	service := NewService(client)

	ctx := context.Background()
	req := SearchRequest{
		Index:     "test-index",
		Query:     "*",
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
		Page:      1,
		Size:      10,
		Interval:  "1h",
	}

	resp, err := service.SearchLogs(ctx, req)
	assert.NoError(t, err)
	if resp != nil {
		t.Logf("Total hits: %d", resp.Total)
		t.Logf("Histogram buckets: %d", len(resp.Histogram))
	}
}
