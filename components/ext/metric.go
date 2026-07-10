package ext

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/sirupsen/logrus"
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

func newMetricExt() Ext {
	metricOnce.Do(func() {
		metric = &metricExt{}
		producer, err := kafka.NewDefaultProducer(
			context.TODO(),
			kafka.WithSCRAMSASL(Apollo().KafkaCfg().UserName, Apollo().KafkaCfg().Password),
			kafka.WithTopic(metric.getTopic()),
			kafka.WithAddress(metric.getBroker()...),
		)
		if err != nil {
			logrus.WithField("err", err).Panicf("init metric builder failed")
		}

		tsdb.Init(producer)
		metric.producer = producer
	})
	return metric
}

func (*metricExt) Init(_ context.Context) error {
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
