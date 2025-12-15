package kafka

import (
	"time"

	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
)

type Options struct {
	addresses    []string
	topic        string
	sasl         sasl.Mechanism
	writeTimeout time.Duration
	readTimeout  time.Duration
}

type Option func(*Options)

func newOptions(opts ...Option) *Options {
	var options = &Options{
		addresses:    nil,
		topic:        "",
		sasl:         nil,
		writeTimeout: 10 * time.Second,
		readTimeout:  10 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithPlainSASL(userName, password string) Option {
	return func(options *Options) {
		options.sasl = &plain.Mechanism{
			Username: userName,
			Password: password,
		}
	}
}

func WithSCRAMSASL(userName, password string) Option {
	return func(options *Options) {
		mechanism, err := scram.Mechanism(scram.SHA512, userName, password)
		if err != nil {
			logrus.Panic(err)
		}
		options.sasl = mechanism
	}
}

func WithTopic(topic string) Option {
	return func(options *Options) {
		options.topic = topic
	}
}

func WithAddress(addr ...string) Option {
	return func(options *Options) {
		options.addresses = append(options.addresses, addr...)
	}
}
