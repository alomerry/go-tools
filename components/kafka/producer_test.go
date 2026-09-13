package kafka

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/env"
	"github.com/segmentio/kafka-go"
)

func TestCreateMessage(t *testing.T) {
	// 与 admin_client_test 同款探测：无本地 broker 时跳过集成测试
	for _, addr := range []string{"localhost:9391", "localhost:9392", "localhost:9393"} {
		if conn, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
			_ = conn.Close()
			break
		}
		t.Skipf("kafka broker %s unreachable, skip integration test", addr)
	}

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
