package tsdb

/*
// NewMetric creates a new Metric client for writing to InfluxDB.
// It uses environment variables as defaults if options are not provided.
//
// Example usage:
//
//	metric, err := tsdb.NewMetric(ctx,
//		def.WithEndpoint("http://localhost:8086"),
//		def.WithOrg("my-org"),
//		def.WithBucket("my-bucket"),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer metric.Close()
//
//	// Write a simple point
//	err = metric.LogPoint("my-bucket", "temperature", map[string]string{"sensor": "A1"}, map[string]any{"value": 23.5})
//
//	// Use builder pattern
//	point := def.NewPointBuilder("temperature").
//		Tag("sensor", "A1").
//		Tag("location", "room1").
//		Field("value", 23.5).
//		Field("unit", "celsius").
//		Build()
//	err = metric.LogPoints("my-bucket", []def.Point{point})
//
//	// Use convenience methods
//	err = metric.Counter("my-bucket", "request_count", map[string]string{"method": "GET"}, 1)
//	err = metric.Gauge("my-bucket", "cpu_usage", map[string]string{"host": "server1"}, 45.2)
func NewMetric(ctx context.Context, options ...def.Option) (def.Metric, error) {
	return internal.NewDefaultCat(ctx, options...)
}

// NewMetricWithDefaults creates a new Metric client using environment variables as defaults.
// This is a convenience function that automatically uses INFLUXDB_ENDPOINT, INFLUXDB_ORG, and INFLUXDB_TOKEN.
func NewMetricWithDefaults(ctx context.Context, options ...def.Option) (def.Metric, error) {
	opts := []def.Option{
		def.WithEndpoint(influxdb.GetEndpoint()),
		def.WithOrg(influxdb.GetOrg()),
	}
	opts = append(opts, options...)
	return NewMetric(ctx, opts...)
}

// NewPointBuilder creates a new point builder for constructing data points with a fluent API.
//
// Example:
//
//	point := tsdb.NewPointBuilder("temperature").
//		Tag("sensor", "A1").
//		Field("value", 23.5).
//		Time(time.Now()).
//		Build()
func NewPointBuilder(measurement string) *def.PointBuilder {
	return def.NewPointBuilder(measurement)
}*/
