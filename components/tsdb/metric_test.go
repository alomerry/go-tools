package tsdb

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb/internal"
	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestAddMetricMsg(t *testing.T) {
	producer, err := kafka.NewDefaultProducer(
		context.TODO(),
		kafka.WithAddress("localhost:9391", "localhost:9392", "localhost:9393"),
		kafka.WithTopic("test-topic"),
		kafka.WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)

	assert.Nil(t, err)

	Init(WithProducer(producer))
	m := NewMetric(WithMetric("test"), WithTag("type", "6666")).(*metric)
	m.LVals["cnt"] = 1

	err = internal.Write(m)
	assert.Nil(t, err)
}
