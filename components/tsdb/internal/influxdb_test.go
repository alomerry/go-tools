package internal

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/cons/tsdb"
	"github.com/stretchr/testify/assert"
)

func TestReadMetric(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, "", token)

	assert.NoError(t, err)
	defer client.Close()

	options := append([]func(*def.TsdbQueryOptions){},
		def.WithBucket("homelab"),
		def.WithMeasurement("cpu.usage"),
		def.WithFields("usage", "cnt"),
		def.WithGroup("service"),
		def.WithTag("service", tsdb.OpEqual, "homelab-backend-account"),
	)

	client.Query(ctx, options...)

}

func TestDefault_LogPoint(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"
	measurement := "measurement1"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, "", token)

	assert.NoError(t, err)
	defer client.Close()

	t.Run("writes point successfully", func(t *testing.T) {
		err := client.LogPoint(ctx, bucket, measurement,
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("returns error for empty bucket", func(t *testing.T) {
		err := client.LogPoint(ctx, "", measurement,
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.Error(t, err)
	})
}

func TestDefault_LogPointWithTime(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, "", token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("writes point with custom time", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		err := client.LogPointWithTime(ctx, "test-bucket", "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
			customTime,
		)
		assert.NoError(t, err)
	})
}

func TestNewDefaultCat(t *testing.T) {
	t.Run("creates client with options", func(t *testing.T) {
		endpoint := os.Getenv("INFLUXDB_ENDPOINT")
		org := os.Getenv("INFLUXDB_ORG")
		token := os.Getenv("INFLUXDB_TOKEN")

		ctx := context.Background()
		client, err := NewInfluxdbClient(ctx, org, endpoint, "", token)

		// May fail if InfluxDB is not running
		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, client)
		if client != nil {
			defer client.Close()
		}
	})

	t.Run("validates endpoint", func(t *testing.T) {
		org := os.Getenv("INFLUXDB_ORG")
		token := os.Getenv("INFLUXDB_TOKEN")

		ctx := context.Background()
		_, err := NewInfluxdbClient(ctx, org, "", "", token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint")
	})

	t.Run("validates org", func(t *testing.T) {
		endpoint := os.Getenv("INFLUXDB_ENDPOINT")
		token := os.Getenv("INFLUXDB_TOKEN")

		ctx := context.Background()
		_, err := NewInfluxdbClient(ctx, "", endpoint, "", token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org")
	})

}

func TestDefaultCat_LogPoints(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("writes multiple points", func(t *testing.T) {
		points := []def.Point{
			def.NewPointBuilder("test_measurement").
				Tag("tag1", "value1").
				Field("field1", 1.0).
				Build(),
			def.NewPointBuilder("test_measurement").
				Tag("tag1", "value2").
				Field("field1", 2.0).
				Build(),
			def.NewPointBuilder("test_measurement").
				Tag("tag1", "value3").
				Field("field1", 3.0).
				Build(),
		}
		err := client.LogPoints(ctx, "test-bucket", points)
		assert.NoError(t, err)
	})

	t.Run("handles empty points array", func(t *testing.T) {
		err := client.LogPoints(ctx, "test-bucket", []def.Point{})
		assert.NoError(t, err)
	})

	t.Run("returns error for empty bucket", func(t *testing.T) {
		points := []def.Point{
			def.NewPointBuilder("test_measurement").
				Field("field1", 1.0).
				Build(),
		}
		err := client.LogPoints(ctx, "", points)
		assert.Error(t, err)
	})
}

func TestDefaultCat_LogPointToDefault(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("writes to default bucket", func(t *testing.T) {
		err := client.LogPointToDefault(ctx, "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("returns error when default bucket not set", func(t *testing.T) {
		clientNoBucket, err := NewInfluxdbClient(ctx, org, endpoint, "", token)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}
		defer clientNoBucket.Close()

		err = clientNoBucket.LogPointToDefault(ctx, "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.Error(t, err)
	})
}

func TestDefaultCat_MetricTypes(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("writes counter metric", func(t *testing.T) {
		err := client.Counter(ctx, "test-bucket", "test_counter",
			map[string]string{"method": "GET"},
			1.0,
		)
		assert.NoError(t, err)
	})

	t.Run("writes gauge metric", func(t *testing.T) {
		err := client.Gauge(ctx, "test-bucket", "test_gauge",
			map[string]string{"host": "server1"},
			45.2,
		)
		assert.NoError(t, err)
	})

	t.Run("writes histogram metric", func(t *testing.T) {
		err := client.Histogram(ctx, "test-bucket", "test_histogram",
			map[string]string{"operation": "read"},
			123.45,
		)
		assert.NoError(t, err)
	})

	t.Run("writes summary metric", func(t *testing.T) {
		err := client.Summary(ctx, "test-bucket", "test_summary",
			map[string]string{"service": "api"},
			67.89,
		)
		assert.NoError(t, err)
	})
}

func TestDefaultCat_Ping(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("ping succeeds when InfluxDB is available", func(t *testing.T) {
		err := client.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestDefaultCat_Close(t *testing.T) {
	t.Run("closes without error", func(t *testing.T) {
		endpoint := os.Getenv("INFLUXDB_ENDPOINT")
		org := os.Getenv("INFLUXDB_ORG")
		token := os.Getenv("INFLUXDB_TOKEN")

		if endpoint == "" || org == "" || token == "" {
			t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
		}

		bucket := "homelab"

		ctx := context.Background()
		client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		err = client.Close()
		assert.NoError(t, err)
	})

	t.Run("handles nil client gracefully", func(t *testing.T) {
		// Create a client with nil client to test Close doesn't panic
		client := &influxdbClient{}
		// This should not panic even with nil client
		err := client.Close()
		assert.NoError(t, err)
	})
}

// TestDefaultCat_WriteAPICaching tests that write APIs are cached
// This is tested indirectly through multiple writes to the same bucket
func TestDefaultCat_WriteAPICaching(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"

	ctx := context.Background()
	client, err := NewInfluxdbClient(ctx, org, endpoint, bucket, token)

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("multiple writes to same bucket work", func(t *testing.T) {
		// Multiple writes to the same bucket should use cached API
		err1 := client.LogPoint(ctx, "test-bucket", "test1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 1.0},
		)
		assert.NoError(t, err1)

		err2 := client.LogPoint(ctx, "test-bucket", "test2",
			map[string]string{"tag1": "value2"},
			map[string]any{"field1": 2.0},
		)
		assert.NoError(t, err2)
	})

	t.Run("writes to different buckets work", func(t *testing.T) {
		err1 := client.LogPoint(ctx, "bucket1", "test1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 1.0},
		)
		assert.NoError(t, err1)

		err2 := client.LogPoint(ctx, "bucket2", "test2",
			map[string]string{"tag1": "value2"},
			map[string]any{"field1": 2.0},
		)
		assert.NoError(t, err2)
	})
}
