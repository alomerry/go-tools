package def

import (
	"context"
	"time"
)

type Option func(*Meta)

type Meta struct {
	Org      string // organization
	Endpoint string
	Token    string // optional, can be set via environment variable
	Bucket   string // default bucket
}

func WithEndpoint(endpoint string) Option {
	return func(meta *Meta) {
		meta.Endpoint = endpoint
	}
}

func WithOrg(org string) Option {
	return func(meta *Meta) {
		meta.Org = org
	}
}

func WithToken(token string) Option {
	return func(meta *Meta) {
		meta.Token = token
	}
}

func WithBucket(bucket string) Option {
	return func(meta *Meta) {
		meta.Bucket = bucket
	}
}

// Point represents a single data point to be written to InfluxDB
type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	Time        time.Time
}

// Metric is the main interface for writing metrics to InfluxDB
type Metric interface {
	// LogPoint writes a single point to the specified bucket
	LogPoint(bucket, measurement string, tags map[string]string, fields map[string]any) error

	// LogPointWithTime writes a single point with a specific timestamp
	LogPointWithTime(bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error

	// LogPoints writes multiple points in a batch (more efficient)
	LogPoints(bucket string, points []Point) error

	// LogPointToDefault writes to the default bucket if configured
	LogPointToDefault(measurement string, tags map[string]string, fields map[string]any) error

	// Counter increments a counter metric
	Counter(bucket, measurement string, tags map[string]string, value float64) error

	// Gauge sets a gauge metric value
	Gauge(bucket, measurement string, tags map[string]string, value float64) error

	// Histogram records a histogram value
	Histogram(bucket, measurement string, tags map[string]string, value float64) error

	// Summary records a summary value
	Summary(bucket, measurement string, tags map[string]string, value float64) error

	// Ping checks if the InfluxDB connection is healthy
	Ping(ctx context.Context) error

	// Close closes the client connection and releases resources
	Close() error
}

// PointBuilder helps build points with a fluent API
type PointBuilder struct {
	measurement string
	tags        map[string]string
	fields      map[string]any
	time        time.Time
}

// NewPointBuilder creates a new point builder
func NewPointBuilder(measurement string) *PointBuilder {
	return &PointBuilder{
		measurement: measurement,
		tags:        make(map[string]string),
		fields:      make(map[string]any),
		time:        time.Now(),
	}
}

// Tag adds a tag to the point
func (pb *PointBuilder) Tag(key, value string) *PointBuilder {
	pb.tags[key] = value
	return pb
}

// Tags adds multiple tags to the point
func (pb *PointBuilder) Tags(tags map[string]string) *PointBuilder {
	for k, v := range tags {
		pb.tags[k] = v
	}
	return pb
}

// Field adds a field to the point
func (pb *PointBuilder) Field(key string, value any) *PointBuilder {
	pb.fields[key] = value
	return pb
}

// Fields adds multiple fields to the point
func (pb *PointBuilder) Fields(fields map[string]any) *PointBuilder {
	for k, v := range fields {
		pb.fields[k] = v
	}
	return pb
}

// Time sets the timestamp for the point
func (pb *PointBuilder) Time(t time.Time) *PointBuilder {
	pb.time = t
	return pb
}

// Build creates the Point from the builder
func (pb *PointBuilder) Build() Point {
	return Point{
		Measurement: pb.measurement,
		Tags:        pb.tags,
		Fields:      pb.fields,
		Time:        pb.time,
	}
}
