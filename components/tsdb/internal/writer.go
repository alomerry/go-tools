package internal

import (
	"context"
	"sync"
	"time"

	"github.com/alomerry/go-tools/components/kafka"
	kafka2 "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var (
	pool     chan Serializer
	producer *kafka.Producer
	once     = sync.Once{}
)

type Serializer interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

func AsyncWrite(s Serializer) {
	pool <- s
}

func Write(s Serializer) error {
	data, err := s.Encode()
	if err != nil {
		return err
	}
	err = producer.Write(context.TODO(), kafka2.Message{
		Key: []byte(cast.ToString(time.Now().Unix())), Value: data,
	})
	if err != nil {
		logrus.WithField("err", err).Errorf("write metric to kafka error")
		return err
	}
	return nil
}

func InitMetricWriter(p *kafka.Producer) {
	once.Do(func() {
		producer = p
		pool = make(chan Serializer, 1000)
		go run()
	})
}

func run() {
	for {
		select {
		case it := <-pool:
			data, err := it.Encode()
			if err != nil {
				logrus.WithField("err", err.Error()).Errorf("metric encode failed")
				continue
			}

			err = producer.Write(context.TODO(), kafka2.Message{
				Key: []byte(cast.ToString(time.Now().Unix())), Value: data,
			})
			if err != nil {
				logrus.WithField("err", err).Errorf("write metric to kafka error")
			}
		}
	}
}
