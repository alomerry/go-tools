package def

import (
	"context"
	"time"
)

type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	Time        time.Time
}

type TsdbWriter interface {
	LogPoint(ctx context.Context, bucket, measurement string, tags map[string]string, fields map[string]any) error
	LogPointWithTime(ctx context.Context, bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error
	LogPoints(ctx context.Context, bucket string, points []Point) error
	LogPointToDefault(ctx context.Context, measurement string, tags map[string]string, fields map[string]any) error
	Counter(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error
	Gauge(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error
	Histogram(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error
	Summary(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error
	Ping(ctx context.Context) error
	Close() error
}

type PointBuilder struct {
	measurement string
	tags        map[string]string
	fields      map[string]any
	time        time.Time
}

func NewPointBuilder(measurement string) *PointBuilder {
	return &PointBuilder{
		measurement: measurement,
		tags:        make(map[string]string),
		fields:      make(map[string]any),
		time:        time.Now(),
	}
}

func (pb *PointBuilder) Tag(key, value string) *PointBuilder {
	pb.tags[key] = value
	return pb
}

func (pb *PointBuilder) Tags(tags map[string]string) *PointBuilder {
	for k, v := range tags {
		pb.tags[k] = v
	}
	return pb
}

func (pb *PointBuilder) Field(key string, value any) *PointBuilder {
	pb.fields[key] = value
	return pb
}

func (pb *PointBuilder) Fields(fields map[string]any) *PointBuilder {
	for k, v := range fields {
		pb.fields[k] = v
	}
	return pb
}

func (pb *PointBuilder) Time(t time.Time) *PointBuilder {
	pb.time = t
	return pb
}

func (pb *PointBuilder) Build() Point {
	return Point{
		Measurement: pb.measurement,
		Tags:        pb.tags,
		Fields:      pb.fields,
		Time:        pb.time,
	}
}
