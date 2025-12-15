package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/env"
	"github.com/segmentio/kafka-go"
)

func TestCreateMessage(t *testing.T) {
	var (
		topic  = "test-topic"
		ctx    = context.TODO()
		client = getConn(ctx)
		msg    = `{"metric":"cpu_usage","tags":{"host":"server1","region":"us-east-1"},"lVals":{"ts":1718099200},"fVals":{"value":0.82}}`
	)

	producer, _ := NewDefaultProducer(
		ctx,
		WithAddress("localhost:9391", "localhost:9392", "localhost:9393"),
		WithTopic(topic),
		WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)

	err := producer.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(time.Now().String()), Value: []byte(msg),
	})

	if err != nil {
		t.Fatal("failed to write messages:", err)
	}

	_ = client.Close()
}
