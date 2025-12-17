package tsdb

import (
	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/components/tsdb/internal"
)

type metric struct {
	Measurement string             `json:"metric"`
	Tags        map[string]string  `json:"tags"`
	Bucket      string             `json:"bucket"`
	LVals       map[string]int64   `json:"lVals"`
	FVals       map[string]float64 `json:"fVals"`
}

func NewMetric(opts ...func(*metric)) def.Metric {
	m := &metric{
		Tags:  make(map[string]string),
		LVals: make(map[string]int64),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func WithMetric(measurement string) func(*metric) {
	return func(m *metric) {
		m.Measurement = measurement
	}
}

func WithTag(k, v string) func(*metric) {
	return func(m *metric) {
		if k == "" || v == "" {
			return
		}
		if m.Tags == nil {
			m.Tags = make(map[string]string)
		}
		m.Tags[k] = v
	}
}

func WithBucket(bucket string) func(*metric) {
	return func(m *metric) {
		m.Bucket = bucket
	}
}

func (m *metric) LogForCnt(count int64) {
	m.LVals["cnt"] = count
	internal.AsyncWrite(m)
}
