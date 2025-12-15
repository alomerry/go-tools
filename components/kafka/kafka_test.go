package kafka

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/static/env"
	"github.com/stretchr/testify/assert"
)

func TestCreateTopic(t *testing.T) {
	var (
		ctx    = context.TODO()
		client = getConn(ctx)
	)
	err := client.CreateTopic(ctx, "test-topic", 1, 1)
	assert.Nil(t, err)
	_ = client.Close()
}

func TestDeleteTopics(t *testing.T) {
	var (
		ctx    = context.TODO()
		client = getConn(ctx)
	)
	err := client.DeleteTopics(ctx, "test-topic")
	assert.Nil(t, err)
	_ = client.Close()
}

func TestListTopics(t *testing.T) {
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
	return NewKafkaClient(
		ctx,
		WithAddress("127.0.0.1:9393"),
		WithSCRAMSASL(env.GetKafkaUserName(), env.GetKafkaPassword()),
	)
}
