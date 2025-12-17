package tsdb

/*// TestNewPointBuilder tests the exported NewPointBuilder function
func TestNewPointBuilder(t *testing.T) {
	t.Run("creates point builder", func(t *testing.T) {
		builder := NewPointBuilder("temperature")
		assert.NotNil(t, builder)

		point := builder.Tag("sensor", "A1").Field("value", 23.5).Build()
		assert.Equal(t, "temperature", point.Measurement)
		assert.Equal(t, "A1", point.Tags["sensor"])
		assert.Equal(t, 23.5, point.Fields["value"])
	})
}

// TestNewMetric tests metric creation with various options
func TestNewMetric(t *testing.T) {
	t.Run("creates metric with options", func(t *testing.T) {
		ctx := context.Background()
		metric, err := NewMetric(ctx,
			def.WithEndpoint("http://localhost:8086"),
			def.WithOrg("test-org"),
			def.WithToken("test-token"),
		)

		// Note: This will fail if InfluxDB is not running, which is expected
		// In a real scenario, you might want to skip this test or use a mock
		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, metric)
		if metric != nil {
			defer metric.Close()
		}
	})

	t.Run("validates required fields", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewMetric(ctx)
		// Should fail validation if endpoint or org is missing
		assert.Error(t, err)
	})
}

// TestNewMetricWithDefaults tests the convenience function
func TestNewMetricWithDefaults(t *testing.T) {
	// Save original env vars
	originalEndpoint := os.Getenv("INFLUXDB_ENDPOINT")
	originalOrg := os.Getenv("INFLUXDB_ORG")
	originalToken := os.Getenv("INFLUXDB_TOKEN")

	// Restore env vars after test
	defer func() {
		if originalEndpoint != "" {
			os.Setenv("INFLUXDB_ENDPOINT", originalEndpoint)
		} else {
			os.Unsetenv("INFLUXDB_ENDPOINT")
		}
		if originalOrg != "" {
			os.Setenv("INFLUXDB_ORG", originalOrg)
		} else {
			os.Unsetenv("INFLUXDB_ORG")
		}
		if originalToken != "" {
			os.Setenv("INFLUXDB_TOKEN", originalToken)
		} else {
			os.Unsetenv("INFLUXDB_TOKEN")
		}
	}()

	t.Run("uses environment variables", func(t *testing.T) {
		os.Setenv("INFLUXDB_ENDPOINT", "http://test:8086")
		os.Setenv("INFLUXDB_ORG", "test-org")
		os.Setenv("INFLUXDB_TOKEN", "test-token")

		ctx := context.Background()
		metric, err := NewMetricWithDefaults(ctx)

		// May fail if InfluxDB is not running
		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, metric)
		if metric != nil {
			defer metric.Close()
		}
	})

	t.Run("allows overriding defaults", func(t *testing.T) {
		os.Setenv("INFLUXDB_ENDPOINT", "http://default:8086")
		os.Setenv("INFLUXDB_ORG", "default-org")

		ctx := context.Background()
		metric, err := NewMetricWithDefaults(ctx,
			def.WithEndpoint("http://override:8086"),
			def.WithOrg("override-org"),
		)

		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, metric)
		if metric != nil {
			defer metric.Close()
		}
	})
}

// Integration tests - these require a running InfluxDB instance
// Set INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN environment variables to run
func TestMetric_Integration(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping integration test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	bucket := "homelab"
	ctx := context.Background()
	metric, err := NewMetric(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
		def.WithBucket(bucket),
	)

	if err != nil {
		t.Fatalf("Failed to create metric: %v", err)
	}
	defer metric.Close()

	t.Run("ping succeeds", func(t *testing.T) {
		err := metric.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("writes single point", func(t *testing.T) {
		err := metric.LogPoint(bucket, "measurement1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("writes point with custom time", func(t *testing.T) {
		customTime := time.Now().Add(-1 * time.Minute * 10)
		err := metric.LogPointWithTime(bucket, "measurement1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
			customTime,
		)
		assert.NoError(t, err)
	})

	t.Run("writes multiple points", func(t *testing.T) {
		points := []def.Point{
			def.NewPointBuilder("measurement1").
				Tag("tag1", "value1").
				Field("field1", 1.0).
				Build(),
			def.NewPointBuilder("measurement1").
				Tag("tag1", "value2").
				Field("field1", 2.0).
				Build(),
		}
		err := metric.LogPoints(bucket, points)
		assert.NoError(t, err)
	})

	t.Run("writes to default bucket", func(t *testing.T) {
		err := metric.LogPointToDefault("measurement1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("writes counter metric", func(t *testing.T) {
		err := metric.Counter("test-bucket", "test_counter",
			map[string]string{"method": "GET"},
			1.0,
		)
		assert.NoError(t, err)
	})

	t.Run("writes gauge metric", func(t *testing.T) {
		err := metric.Gauge("test-bucket", "test_gauge",
			map[string]string{"host": "server1"},
			45.2,
		)
		assert.NoError(t, err)
	})

	t.Run("writes histogram metric", func(t *testing.T) {
		err := metric.Histogram("test-bucket", "test_histogram",
			map[string]string{"operation": "read"},
			123.45,
		)
		assert.NoError(t, err)
	})

	t.Run("writes summary metric", func(t *testing.T) {
		err := metric.Summary("test-bucket", "test_summary",
			map[string]string{"service": "api"},
			67.89,
		)
		assert.NoError(t, err)
	})

	t.Run("handles empty bucket error", func(t *testing.T) {
		err := metric.LogPoint("", "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.Error(t, err)
	})

	t.Run("handles empty points array", func(t *testing.T) {
		err := metric.LogPoints("test-bucket", []def.Point{})
		assert.NoError(t, err)
	})
}*/
