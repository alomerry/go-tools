package ext

import (
	"context"
	"time"

	"github.com/alomerry/go-tools/components/monitor"
	"github.com/alomerry/go-tools/components/tsdb"
	monitor2 "github.com/alomerry/go-tools/model/monitor"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
)

func init() {
	Register(cons.ExtSysMonitor, newSysMonitorExt)
}

type sysMonitorExt struct {
	ctx     context.Context
	monitor monitor.SystemMonitor
}

func newSysMonitorExt() Ext {
	s := &sysMonitorExt{
		ctx: context.TODO(),
	}
	s.monitor = monitor.NewSystemMonitor(
		monitor.WithContext(s.ctx),
		monitor.WithInterval(5*time.Second),
		// monitor.WithCategory(monitor.SystemMonitorCategoryDocker),
		monitor.WithCallback(s.fn))
	return s
}

func (s *sysMonitorExt) Init(_ context.Context) error {
	s.monitor.Watch()
	return nil
}

func (*sysMonitorExt) fn(stats *monitor2.SystemStats) error {
	tsdb.NewMetric(
		tsdb.WithMetric("cpu.usage"),
		tsdb.WithField("usage", stats.CpuUsage),
		tsdb.WithField("system", stats.SystemUsage),
		tsdb.WithField("user", stats.UserUsage),
		tsdb.WithTag("service", env.GetService()),
		tsdb.WithTag("container", stats.Name),
	).LogForCnt()

	tsdb.NewMetric(
		tsdb.WithMetric("memory.usage"),
		tsdb.WithField("total", stats.TotalMemory),
		tsdb.WithField("used", stats.UsedMemory),
		tsdb.WithField("rss", stats.RssMemory),
		tsdb.WithField("cache", stats.CachedMemory),
		tsdb.WithField("usage", stats.MemoryUsage),
		tsdb.WithTag("service", env.GetService()),
		tsdb.WithTag("container", stats.Name),
	).LogForCnt()

	return nil
}
