package kafka

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

// brokerReachable 探测本地 broker 是否可达；无 broker 环境下跳过集成测试
func brokerReachable(t *testing.T, addr string) {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Skipf("kafka broker %s unreachable, skip integration test: %v", addr, err)
	}
	_ = conn.Close()
}

func TestCreateTopic(t *testing.T) {
	brokerReachable(t, "127.0.0.1:9393")
	var (
		ctx    = context.TODO()
		client = getConn(ctx)
	)
	err := client.CreateTopic(ctx, "test-topic", 1, 1)
	assert.Nil(t, err)
	_ = client.Close()
}

func TestDeleteTopics(t *testing.T) {
	brokerReachable(t, "127.0.0.1:9393")
	var (
		ctx    = context.TODO()
		client = getConn(ctx)
	)
	err := client.DeleteTopics(ctx, "test-topic")
	assert.Nil(t, err)
	_ = client.Close()
}

func TestListTopics(t *testing.T) {
	brokerReachable(t, "127.0.0.1:9393")
	var (
		ctx    = context.TODO()
		client = getConn(ctx)
	)
	topics, err := client.ListTopics()
	assert.Nil(t, err)
	t.Log(topics)
	_ = client.Close()
}

func getConn(ctx context.Context) *Client {
	client, err := NewKafkaClient(
		ctx,
		WithAddress("127.0.0.1:9393"),
		WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)
	if err != nil {
		panic(err)
	}
	return client
}
