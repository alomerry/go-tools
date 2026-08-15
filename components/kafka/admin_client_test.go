package kafka

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/env"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBrokerAddr 与既有测试一致的本地 broker 地址。无 broker 环境下跳过集成测试。
const testBrokerAddr = "127.0.0.1:9393"

// newTestAdminClient 探测本地 broker 是否可达；不可达时跳过测试。
func newTestAdminClient(t *testing.T) *AdminClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", testBrokerAddr, 2*time.Second)
	if err != nil {
		t.Skipf("kafka broker %s unreachable, skip integration test: %v", testBrokerAddr, err)
	}
	_ = conn.Close()

	client, err := NewAdminClient(
		WithAddress(testBrokerAddr),
		WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestAdminClientListTopics(t *testing.T) {
	client := newTestAdminClient(t)
	topics, err := client.ListTopics(context.Background())
	require.NoError(t, err)
	t.Logf("topics: %+v", topics)
	for _, tp := range topics {
		assert.NotEmpty(t, tp.Name)
		assert.GreaterOrEqual(t, tp.Partitions, 1)
		assert.GreaterOrEqual(t, tp.ReplicationFactor, 1)
	}
}

func TestAdminClientTopicOffsets(t *testing.T) {
	client := newTestAdminClient(t)
	offsets, err := client.TopicOffsets(context.Background(), nil)
	require.NoError(t, err)
	t.Logf("offsets: %+v", offsets)
	for topic, parts := range offsets {
		assert.NotEmpty(t, topic)
		for _, p := range parts {
			assert.GreaterOrEqual(t, p.HighWatermark, p.LogStartOffset)
		}
	}
}

func TestAdminClientConsumerGroups(t *testing.T) {
	client := newTestAdminClient(t)
	groups, err := client.ListConsumerGroups(context.Background())
	require.NoError(t, err)
	t.Logf("consumer groups: %+v", groups)
	for _, g := range groups {
		assert.NotEmpty(t, g.GroupID)
	}
}

func TestAdminClientConsumerGroupLagEmpty(t *testing.T) {
	client := newTestAdminClient(t)
	// 不存在的组返回空 lag，不报错
	lag, err := client.ConsumerGroupLag(context.Background(), fmt.Sprintf("nonexistent-group-%d", time.Now().UnixNano()))
	require.NoError(t, err)
	assert.Equal(t, int64(0), lag.TotalLag)
	assert.Empty(t, lag.Partitions)
}

// TestAdminClientCreateDeleteTopicAndReadLatest 覆盖：创建 topic → 写入消息 →
// 消息采样读回 → 删除 topic 的完整链路。
func TestAdminClientCreateDeleteTopicAndReadLatest(t *testing.T) {
	client := newTestAdminClient(t)
	ctx := context.Background()

	topic := fmt.Sprintf("test-admin-%d", time.Now().UnixNano())
	err := client.CreateTopic(ctx, topic, 1, 1)
	require.NoError(t, err)
	defer func() { _ = client.DeleteTopic(ctx, topic) }()

	producer, err := NewDefaultProducer(
		ctx,
		WithAddress(testBrokerAddr),
		WithTopic(topic),
		WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)
	require.NoError(t, err)
	defer func() { _ = producer.Close() }()

	err = producer.Write(ctx,
		kafka.Message{Key: []byte("k1"), Value: []byte(`{"msg":"hello-1"}`)},
		kafka.Message{Key: []byte("k2"), Value: []byte(`{"msg":"hello-2"}`)},
	)
	require.NoError(t, err)

	// 等待消息可见（produce 是 RequireAll，通常即刻可见；留 1s 缓冲）
	time.Sleep(500 * time.Millisecond)

	samples, err := client.ReadLatestMessages(ctx, topic, 10)
	require.NoError(t, err)
	require.Len(t, samples, 2)
	assert.Equal(t, []byte(`{"msg":"hello-2"}`), samples[0].Value)
	assert.Equal(t, "k2", samples[0].Key)

	// 删除 topic 后元数据中不再存在
	require.NoError(t, client.DeleteTopic(ctx, topic))
	topics, err := client.ListTopics(ctx)
	require.NoError(t, err)
	for _, tp := range topics {
		assert.NotEqual(t, topic, tp.Name)
	}
}
