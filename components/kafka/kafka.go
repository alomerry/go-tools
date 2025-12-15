package kafka

// 使用 kafka-go 实现查看 topic 元数据、查看 broker 元数据、查看 consumer group 元数据
import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	conn *kafka.Conn
}

func NewKafkaDialer(opts ...Option) (*kafka.Dialer, *Options) {
	var (
		options = new(Options)
	)

	for _, opt := range opts {
		opt(options)
	}

	dialer := &kafka.Dialer{
		Timeout:   30 * time.Second, // Increased timeout for SSH tunnel and nginx proxy
		DualStack: false,            // Disable IPv6 to work with SSH tunnels that only forward IPv4
		KeepAlive: 10 * time.Second, // Enable keep-alive for stable connections through proxy
	}

	if options.sasl != nil {
		dialer.SASLMechanism = options.sasl
	}

	return dialer, options
}

func NewKafkaClient(ctx context.Context, opts ...Option) *Client {
	dialer, options := NewKafkaDialer(opts...)

	var (
		conn *kafka.Conn
		err  error
	)
	conn, err = dialer.DialContext(ctx, "tcp", options.addresses[0])

	if err != nil {
		panic(err.Error())
	}

	return &Client{conn: conn}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) WriteMessages(ctx context.Context, topic string, msg ...kafka.Message) error {
	_, err := c.conn.WriteMessages(msg...)
	return err
}

func (c *Client) CreateTopic(ctx context.Context, topic string, partitions, replication int) error {
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replication,
		},
	}

	err := c.conn.CreateTopics(topicConfigs...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteTopics(ctx context.Context, topics ...string) error {
	err := c.conn.DeleteTopics(topics...)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListTopics() ([]string, error) {
	partitions, err := c.conn.ReadPartitions()
	if err != nil {
		return nil, err
	}

	var (
		topics []string
		mapper = make(map[string]struct{})
	)

	for _, p := range partitions {
		mapper[p.Topic] = struct{}{}
	}

	for t := range mapper {
		topics = append(topics, t)
	}

	return topics, nil
}
