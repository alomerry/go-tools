package monitor

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/sirupsen/logrus"
	net2 "k8s.io/utils/net"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMonitor interface {
	Watch()
}

type option struct {
	ctx      context.Context
	interval time.Duration
	callback func(stats *SystemStats) error
}

func WithContext(ctx context.Context) func(o *option) {
	return func(o *option) {
		o.ctx = ctx
	}
}

func WithInterval(interval time.Duration) func(o *option) {
	return func(o *option) {
		o.interval = interval
	}
}

func WithCallback(callback func(stats *SystemStats) error) func(o *option) {
	return func(o *option) {
		o.callback = callback
	}
}

func NewSystemMonitor(opts ...func(*option)) SystemMonitor {
	var (
		options = new(option)
	)

	for _, opt := range opts {
		opt(options)
	}

	return &systemMonitor{
		opt:  options,
		once: sync.Once{},
	}
}

type systemMonitor struct {
	opt  *option
	once sync.Once
}

func (s *systemMonitor) Watch() {
	s.once.Do(func() {
		go s.run()
	})
}

func (s *systemMonitor) run() {
	ticker := time.NewTicker(s.opt.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.opt.ctx.Done():
			return
		case <-ticker.C:
			stats, err := CollectStats()
			if err != nil {
				logrus.Errorf("收集系统统计信息失败: %v", err)
				continue
			}

			if s.opt.callback == nil {
				continue
			}

			err = s.opt.callback(stats)
			if err != nil {
				logrus.Errorf("回调失败: %v", err)
			}
		}
	}
}

type SystemStats struct {
	Timestamp time.Time

	PhysicalCPU int
	LogicalCPU  int
	CPUUsage    float64

	TotalMemory uint64
	UsedMemory  uint64
	MemoryUsage float64

	DiskUsage map[string]float64
	LoadAvg   [3]float64
}

func CollectStats() (*SystemStats, error) {
	var (
		stats = &SystemStats{
			Timestamp: time.Now(),
			DiskUsage: make(map[string]float64),
		}
		err error
	)

	stats.LogicalCPU, err = cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	stats.PhysicalCPU, err = cpu.Counts(false)
	if err != nil {
		return nil, err
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	if len(cpuPercent) > 0 {
		stats.CPUUsage = cpuPercent[0]
	}

	// 收集内存使用率
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	stats.TotalMemory = v.Total
	stats.UsedMemory = v.Used
	stats.MemoryUsage = v.UsedPercent

	// 收集磁盘使用率（主要分区）
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	for _, part := range partitions {
		// 只监控根目录和主要挂载点
		if part.Mountpoint == "/" || part.Mountpoint == "/home" ||
			part.Mountpoint == "/var" || part.Mountpoint == "/tmp" {
			usage, err := disk.Usage(part.Mountpoint)
			if err == nil {
				stats.DiskUsage[part.Mountpoint] = usage.UsedPercent
				logrus.Infof("磁盘使用率: %s: %.2f%%", part.Mountpoint, usage.UsedPercent)
			}
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		var ips []string
		for _, addr := range i.Addrs {
			// 判断是否是 ipv4
			if net2.IsIPv4CIDRString(addr.Addr) {
				ips = append(ips, addr.Addr)
			}
		}

		if len(ips) == 0 {
			continue
		}
		logrus.Infof("接口: %s (%s)", i.Name, strings.Join(ips, " "))
	}

	avg, err := load.Avg()
	if err != nil {
		return nil, err
	}

	stats.LoadAvg = [3]float64{avg.Load1, avg.Load5, avg.Load15}
	return stats, nil
}
