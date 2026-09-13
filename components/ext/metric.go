package ext

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb"
	"github.com/alomerry/go-tools/static/cons"
)

const (
	metricKafkaTopic   = "METRIC_KAFKA_TOPIC"
	metricKafkaBrokers = "METRIC_KAFKA_BROKERS"
)

var (
	metric     *metricExt
	metricOnce sync.Once
)

func init() {
	Register(cons.ExtMetric, newMetricExt)
}

type metricExt struct {
	producer *kafka.Producer
}

// newMetricExt 仅创建单例结构体；producer 初始化在 Init 中（错误经 LoadExt
// 统一处理），构造函数不再携带初始化副作用。
func newMetricExt() Ext {
	metricOnce.Do(func() {
		metric = &metricExt{}
	})
	return metric
}

// Init 初始化 kafka producer 并注册到 tsdb 异步链路，幂等（已就绪直接返回）。
func (m *metricExt) Init(_ context.Context) error {
	if m.producer != nil {
		return nil
	}

	producer, err := kafka.NewDefaultProducer(
		context.TODO(),
		kafka.WithSCRAMSASL(Apollo().KafkaCfg().UserName, Apollo().KafkaCfg().Password),
		kafka.WithTopic(m.getTopic()),
		kafka.WithAddress(m.getBroker()...),
	)
	if err != nil {
		return fmt.Errorf("init metric producer failed: %w", err)
	}

	tsdb.Init(producer)
	m.producer = producer
	return nil
}

func (*metricExt) getTopic() string {
	v := os.Getenv(metricKafkaTopic)
	if v == "" {
		return Apollo().KafkaCfg().MetricTopic
	}
	return v
}

func (*metricExt) getBroker() []string {
	v := os.Getenv(metricKafkaBrokers)
	if v == "" {
		return Apollo().KafkaCfg().Brokers
	}
	return strings.Split(v, ",")
}
