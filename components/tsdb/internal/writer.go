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

// dropWarnInterval 限频告警间隔：池满/未初始化丢弃点位时，最多按此间隔告警一次。
const dropWarnInterval = 30 * time.Second

// 约束（init-before-use）：pool/producer 仅由 InitMetricWriter（sync.Once）
// 在流量进入前初始化，LoadExt 中 ExtMetric（或 cat 降级触发的 newMetricExt）
// 恒先于业务流量执行，与 AsyncWrite/Write 构成 happens-before，因此热路径
// 上裸读 pool/producer 不加锁；未初始化时 AsyncWrite 按 nil channel 安全丢弃。
var (
	pool     chan Serializer
	producer *kafka.Producer
	once     = sync.Once{}

	dropMu   sync.Mutex
	lastDrop time.Time
)

type Serializer interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

// AsyncWrite 异步投递点位，非阻塞：池满或 writer 未初始化（pool 为 nil）时
// 直接丢弃并限频告警，绝不阻塞业务 goroutine、绝不 panic。
func AsyncWrite(s Serializer) {
	select {
	case pool <- s:
	default:
		warnDrop()
	}
}

func warnDrop() {
	dropMu.Lock()
	defer dropMu.Unlock()
	if time.Since(lastDrop) < dropWarnInterval {
		return
	}
	lastDrop = time.Now()
	logrus.Warnf("metric async pool is full or not initialized, point dropped")
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
