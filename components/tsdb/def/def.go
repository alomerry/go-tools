package def

import (
	"context"
	"time"
)

type Option func(*Meta)

type Meta struct {
	Org      string // service
	Endpoint string
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

type Metric interface {
	LogPoint(bucket, measurement string, tags map[string]string, fields map[string]any) error
	LogPointWithTime(bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error
	Ping(ctx context.Context) error
}
