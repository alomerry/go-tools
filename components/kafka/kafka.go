package kafka

// 使用 kafka-go 实现查看 topic 元数据、查看 broker 元数据、查看 consumer group 元数据
import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	conn   *kafka.Conn
	writer *kafka.Writer
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

func NewKafkaClient(ctx context.Context, opts ...Option) (*Client, error) {
	dialer, options := NewKafkaDialer(opts...)

	if len(options.addresses) == 0 {
		return nil, errors.New("kafka: no broker addresses configured")
	}

	var (
		conn *kafka.Conn
		err  error
	)
	conn, err = dialer.DialContext(ctx, "tcp", options.addresses[0])

	if err != nil {
		return nil, fmt.Errorf("kafka: dial %s failed: %w", options.addresses[0], err)
	}

	client := &Client{conn: conn}
	// conn 无法指定 topic 写入（会按消息落默认分区而非目标 topic），写入走 writer；
	// Writer.Topic 留空，topic 由 WriteMessages 的消息自带。
	transport := &kafka.Transport{
		DialTimeout: dialer.Timeout,
		Dial:        (&net.Dialer{Timeout: dialer.Timeout}).DialContext,
	}
	if options.sasl != nil {
		transport.SASL = options.sasl
	}
	client.writer = &kafka.Writer{
		Addr:         kafka.TCP(options.addresses...),
		Balancer:     &kafka.ReferenceHash{},
		WriteTimeout: options.writeTimeout,
		ReadTimeout:  options.readTimeout,
		MaxAttempts:  3,
		RequiredAcks: kafka.RequireAll,
		Transport:    transport,
	}

	return client, nil
}

func (c *Client) Close() error {
	err := c.conn.Close()
	if c.writer != nil {
		if werr := c.writer.Close(); err == nil {
			err = werr
		}
	}
	return err
}

func (c *Client) WriteMessages(ctx context.Context, topic string, msg ...kafka.Message) error {
	if c.writer == nil {
		return errors.New("kafka: writer unavailable, no broker addresses configured")
	}
	messages := make([]kafka.Message, 0, len(msg))
	for _, m := range msg {
		if m.Topic == "" {
			m.Topic = topic
		}
		messages = append(messages, m)
	}
	return c.writer.WriteMessages(ctx, messages...)
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
