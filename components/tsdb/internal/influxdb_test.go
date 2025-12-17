package internal

import (
	"encoding/json"
	"os"
	"testing"
)

//     [{"c": 391,"sf": 1,"sl": 0,"sc": 197},{"c": 391,"sf": 1,"sl": 0,"sc": 197}]
// =>  [{"index": 0, "c": 391,"sf": 1,"sl": 0,"sc": 197},{"index": 1, "c": 391,"sf": 1,"sl": 0,"sc": 197}]

type File struct {
	Lines  []Data `json:"lines"`
	LineNo int    `json:"lineNo"`
}

type Data struct {
	Index int `json:"index"`
	C     int `json:"c"`
	Sf    int `json:"sf"`
	Sl    int `json:"sl"`
	Sc    int `json:"sc"`
}

func TestJson(t *testing.T) {
	var (
		files    []File
		resFiles []File
	)
	var data, err = os.ReadFile("/Users/alomerry/Downloads/Untitled-3.json")

	if err != nil {
		panic(err)
	}

	files = Unmarshal(string(data))
	for _, file := range files {
		var dataList []Data
		for index, item := range file.Lines {
			dataList = append(dataList, Data{Index: index, C: item.C, Sf: item.Sf, Sl: item.Sl, Sc: item.Sc})
		}

		resFiles = append(resFiles, File{Lines: dataList, LineNo: file.LineNo})
	}

	// write to file
	res, err := json.Marshal(resFiles)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("/Users/alomerry/workspace/go/go-tools/output/a.json", res, 0644)
	if err != nil {
		panic(err)
	}
}

func Unmarshal(data string) []File {
	var dataList []File
	err := json.Unmarshal([]byte(data), &dataList)
	if err != nil {
		panic(err)
	}
	return dataList
}

/*func TestNewDefaultCat(t *testing.T) {
	t.Run("creates cat with options", func(t *testing.T) {
		ctx := context.Background()
		cat, err := NewDefaultCat(ctx,
			def.WithEndpoint("http://localhost:8086"),
			def.WithOrg("test-org"),
			def.WithToken("test-token"),
		)

		// May fail if InfluxDB is not running
		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, cat)
		if cat != nil {
			defer cat.Close()
		}
	})

	t.Run("validates endpoint", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewDefaultCat(ctx,
			def.WithOrg("test-org"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint")
	})

	t.Run("validates org", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewDefaultCat(ctx,
			def.WithEndpoint("http://localhost:8086"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org")
	})

	t.Run("uses environment variables as defaults", func(t *testing.T) {
		// Save original env vars
		originalEndpoint := os.Getenv("INFLUXDB_ENDPOINT")
		originalOrg := os.Getenv("INFLUXDB_ORG")
		originalToken := os.Getenv("INFLUXDB_TOKEN")

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

		os.Setenv("INFLUXDB_ENDPOINT", "http://test:8086")
		os.Setenv("INFLUXDB_ORG", "test-org")
		os.Setenv("INFLUXDB_TOKEN", "test-token")

		ctx := context.Background()
		cat, err := NewDefaultCat(ctx)

		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, cat)
		if cat != nil {
			defer cat.Close()
		}
	})

	t.Run("uses token from option over environment", func(t *testing.T) {
		// Save original env vars
		originalToken := os.Getenv("INFLUXDB_TOKEN")
		defer func() {
			if originalToken != "" {
				os.Setenv("INFLUXDB_TOKEN", originalToken)
			} else {
				os.Unsetenv("INFLUXDB_TOKEN")
			}
		}()

		os.Setenv("INFLUXDB_TOKEN", "env-token")

		ctx := context.Background()
		cat, err := NewDefaultCat(ctx,
			def.WithEndpoint("http://localhost:8086"),
			def.WithOrg("test-org"),
			def.WithToken("option-token"),
		)

		if err != nil {
			t.Logf("Expected error when InfluxDB is not available: %v", err)
			return
		}

		assert.NotNil(t, cat)
		if cat != nil {
			defer cat.Close()
		}
		// Token from option should be used (we can't directly verify this, but it's tested indirectly)
	})
}

func TestDefaultCat_LogPoint(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("writes point successfully", func(t *testing.T) {
		err := cat.LogPoint("test-bucket", "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("returns error for empty bucket", func(t *testing.T) {
		err := cat.LogPoint("", "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.Error(t, err)
	})
}

func TestDefaultCat_LogPointWithTime(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("writes point with custom time", func(t *testing.T) {
		customTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		err := cat.LogPointWithTime("test-bucket", "test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
			customTime,
		)
		assert.NoError(t, err)
	})
}

func TestDefaultCat_LogPoints(t *testing.T) {
	endpoint := os.Getenv("INFLUXDB_ENDPOINT")
	org := os.Getenv("INFLUXDB_ORG")
	token := os.Getenv("INFLUXDB_TOKEN")

	if endpoint == "" || org == "" || token == "" {
		t.Skip("Skipping test: INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN must be set")
	}

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

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
		err := cat.LogPoints("test-bucket", points)
		assert.NoError(t, err)
	})

	t.Run("handles empty points array", func(t *testing.T) {
		err := cat.LogPoints("test-bucket", []def.Point{})
		assert.NoError(t, err)
	})

	t.Run("returns error for empty bucket", func(t *testing.T) {
		points := []def.Point{
			def.NewPointBuilder("test_measurement").
				Field("field1", 1.0).
				Build(),
		}
		err := cat.LogPoints("", points)
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

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
		def.WithBucket("test-bucket"),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("writes to default bucket", func(t *testing.T) {
		err := cat.LogPointToDefault("test_measurement",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 42.0},
		)
		assert.NoError(t, err)
	})

	t.Run("returns error when default bucket not set", func(t *testing.T) {
		catNoBucket, err := NewDefaultCat(ctx,
			def.WithEndpoint(endpoint),
			def.WithOrg(org),
			def.WithToken(token),
		)
		if err != nil {
			t.Fatalf("Failed to create cat: %v", err)
		}
		defer catNoBucket.Close()

		err = catNoBucket.LogPointToDefault("test_measurement",
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

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("writes counter metric", func(t *testing.T) {
		err := cat.Counter("test-bucket", "test_counter",
			map[string]string{"method": "GET"},
			1.0,
		)
		assert.NoError(t, err)
	})

	t.Run("writes gauge metric", func(t *testing.T) {
		err := cat.Gauge("test-bucket", "test_gauge",
			map[string]string{"host": "server1"},
			45.2,
		)
		assert.NoError(t, err)
	})

	t.Run("writes histogram metric", func(t *testing.T) {
		err := cat.Histogram("test-bucket", "test_histogram",
			map[string]string{"operation": "read"},
			123.45,
		)
		assert.NoError(t, err)
	})

	t.Run("writes summary metric", func(t *testing.T) {
		err := cat.Summary("test-bucket", "test_summary",
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

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("ping succeeds when InfluxDB is available", func(t *testing.T) {
		err := cat.Ping(ctx)
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

		ctx := context.Background()
		cat, err := NewDefaultCat(ctx,
			def.WithEndpoint(endpoint),
			def.WithOrg(org),
			def.WithToken(token),
		)

		if err != nil {
			t.Fatalf("Failed to create cat: %v", err)
		}

		err = cat.Close()
		assert.NoError(t, err)
	})

	t.Run("handles nil client gracefully", func(t *testing.T) {
		// Create a cat with nil client to test Close doesn't panic
		cat := &defaultCat{
			Meta: &def.Meta{},
		}
		// This should not panic even with nil client
		err := cat.Close()
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

	ctx := context.Background()
	cat, err := NewDefaultCat(ctx,
		def.WithEndpoint(endpoint),
		def.WithOrg(org),
		def.WithToken(token),
	)

	if err != nil {
		t.Fatalf("Failed to create cat: %v", err)
	}
	defer cat.Close()

	t.Run("multiple writes to same bucket work", func(t *testing.T) {
		// Multiple writes to the same bucket should use cached API
		err1 := cat.LogPoint("test-bucket", "test1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 1.0},
		)
		assert.NoError(t, err1)

		err2 := cat.LogPoint("test-bucket", "test2",
			map[string]string{"tag1": "value2"},
			map[string]any{"field1": 2.0},
		)
		assert.NoError(t, err2)
	})

	t.Run("writes to different buckets work", func(t *testing.T) {
		err1 := cat.LogPoint("bucket1", "test1",
			map[string]string{"tag1": "value1"},
			map[string]any{"field1": 1.0},
		)
		assert.NoError(t, err1)

		err2 := cat.LogPoint("bucket2", "test2",
			map[string]string{"tag1": "value2"},
			map[string]any{"field1": 2.0},
		)
		assert.NoError(t, err2)
	})
}*/
