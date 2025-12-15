package kafka

import (
	"context"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

// clientId

func NewDefaultProducer(ctx context.Context, opts ...Option) (*Producer, error) {
	var (
		options = new(Options)
	)

	for _, opt := range opts {
		opt(options)
	}

	var (
		timeout = 10 * time.Second
		p       = new(Producer)
	)
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	transport := &kafka.Transport{
		DialTimeout: 10 * time.Second,
		Dial:        dialer.DialContext,
	}

	if options.sasl != nil {
		transport.SASL = options.sasl
	}

	p = &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(options.addresses...),
			Topic:                  options.topic,
			Balancer:               &kafka.ReferenceHash{},
			BatchTimeout:           time.Nanosecond,
			WriteTimeout:           options.writeTimeout,
			ReadTimeout:            options.readTimeout,
			MaxAttempts:            3,
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			Async:                  false,
			Compression:            0,
			Transport:              transport,
		},
	}
	return p, nil
}

func (p *Producer) Write(ctx context.Context, msg ...kafka.Message) error {
	return p.writer.WriteMessages(ctx, msg...)
}
