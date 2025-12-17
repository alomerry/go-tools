package tsdb

import (
	"sync"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb/internal"
)

var (
	initOnce sync.Once
)

func Init(opts ...func(tsdb *tsdbOption)) {
	initOnce.Do(func() {

		option := new(tsdbOption)
		for _, opt := range opts {
			opt(option)
		}

		internal.InitMetricWriter(option.producer)
	})
}

type tsdbOption struct {
	producer *kafka.Producer
}

func WithProducer(producer *kafka.Producer) func(tsdb *tsdbOption) {
	return func(tsdb *tsdbOption) {
		tsdb.producer = producer
	}
}
