package tsdb

import (
	"encoding/json"
	"sync"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb/internal"
)

type Metric interface {
	LogForCnt(...int64)
	LogForVal(map[string]any)
}

type meta struct {
	org      string
	endpoint string
	token    string
	bucket   string
}

var (
	initOnce sync.Once
)

func Init(producer *kafka.Producer) {
	initOnce.Do(func() {
		internal.InitMetricWriter(producer)
	})
}

type metric struct {
	Measurement string            `json:"metric"`
	Tags        map[string]string `json:"tags"`
	Bucket      string            `json:"bucket"`
	Fields      map[string]any    `json:"fields"`
}

func NewMetric(opts ...func(any)) Metric {
	return newMetric(opts...)
}

func newMetric(opts ...func(any)) *metric {
	m := &metric{
		Tags:   make(map[string]string),
		Fields: make(map[string]any),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *metric) LogForCnt(count ...int64) {
	var cnt int64 = 1
	if len(count) > 0 {
		cnt = count[0]
	}
	m.Fields["cnt"] = cnt
	internal.AsyncWrite(m)
}

func (m *metric) LogForVal(mapper map[string]any) {
	for k, v := range mapper {
		withTagOrField(k, v)(m)
	}
	if len(m.Fields) == 0 {
		m.Fields["cnt"] = 1
	}

	internal.AsyncWrite(m)
}

func (m *metric) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (m *metric) Decode(data []byte) error {
	return json.Unmarshal(data, m)
}
