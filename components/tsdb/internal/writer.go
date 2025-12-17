package internal

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/alomerry/go-tools/components/kafka"
	"github.com/alomerry/go-tools/components/tsdb/def"
	kafka2 "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var (
	mb   def.MetricWriter
	once = sync.Once{}
)

func InitMetricWriter(producer *kafka.Producer) {
	once.Do(func() {
		mb = &metricWriter{
			producer: producer,
			ch:       make(chan []byte, 1000),
		}

		go mb.(*metricWriter).run()
	})
}

func AsyncWrite(metric def.Metric) {
	mb.AsyncWrite(metric)
}

func Write(metric def.Metric) error {
	return mb.Write(metric)
}

type metricWriter struct {
	producer *kafka.Producer

	ch chan []byte
}

func (m *metricWriter) AsyncWrite(metric def.Metric) {
	data, err := json.Marshal(metric)
	if err != nil {
		logrus.WithField("err", err).Errorf("marshal metric error")
		return
	}
	m.ch <- data
}

func (m *metricWriter) Write(metric def.Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		logrus.WithField("err", err).Errorf("marshal metric error")
		return err
	}
	err = m.producer.Write(context.TODO(), kafka2.Message{
		Key: []byte(cast.ToString(time.Now().Unix())), Value: data,
	})
	if err != nil {
		logrus.WithField("err", err).Errorf("write metric to kafka error")
		return err
	}
	return nil
}

func (m *metricWriter) run() {
	for {
		select {
		case data := <-m.ch:
			err := m.producer.Write(context.TODO(), kafka2.Message{
				Key: []byte(cast.ToString(time.Now().Unix())), Value: data,
			})
			if err != nil {
				logrus.WithField("err", err).Errorf("write metric to kafka error")
			}
		}
	}
}
