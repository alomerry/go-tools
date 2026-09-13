package ext

import (
	"context"

	"github.com/alomerry/go-tools/components/cat"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/sirupsen/logrus"
)

func init() {
	Register(cons.ExtCat, NewCatExt)
}

type catExt struct {
}

func NewCatExt() Ext {
	return catExt{}
}

func (catExt) Init(ctx context.Context) error {
	domain := env.GetService()
	if len(domain) == 0 {
		domain = "cat"
	}

	// cat 的点位经 ExtMetric 的 Kafka producer 异步上报（tsdb.Init），
	// 而 LoadExt 中 ExtCat 可能先于 ExtMetric 执行，这里显式触发 metric
	// 单例并初始化（newMetricExt 内部 sync.Once 幂等、Init 幂等），保证写入
	// 前 producer 就绪。降级语义：仅 cat 路径容忍 metric 初始化失败（Kafka
	// 不可用时 Warn 后继续，AsyncWrite 未就绪安全丢弃）；metric 自身经
	// LoadExt(ExtMetric) 初始化失败仍会 fail-fast（见 manager.go）。
	if ext := newMetricExt(); ext != nil {
		if err := ext.Init(ctx); err != nil {
			logrus.WithField("err", err).Warn("init metric for cat degraded, continue without kafka producer")
		}
	}

	cat.Init(domain)
	return nil
}
