package pprof

import (
	"context"
	"time"

	"github.com/alomerry/go-tools/components/tsdb"
	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/env/influxdb"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
)

func init() {
	go collectSysInfo()
}

func collectSysInfo() {
	var (
		ctx         = context.Background()
		metric, err = tsdb.NewMetric(ctx, def.WithEndpoint(influxdb.GetEndpoint()), def.WithOrg(influxdb.GetOrg()))
		ticker      = time.NewTicker(time.Second)
		hostInfo, _ = host.Info()
	)

	if err != nil {
		logrus.Panicf("cat init failed: %v", err)
	}

	for {
		<-ticker.C
		vm, _ := mem.VirtualMemory()
		err := metric.LogPoint("system", "mem_usage", map[string]string{
			"hostName": hostInfo.Hostname,
		}, map[string]any{
			"mem_total":   vm.Total,
			"mem_used":    vm.Used,
			"mem_percent": vm.UsedPercent,
		})
		if err != nil {
			logrus.Errorf("cat system info failed: %v", err)
		}

		vc, _ := cpu.Percent(time.Second, false)
		err = metric.LogPoint("system", "cpu_usage", map[string]string{
			"hostName": hostInfo.Hostname,
		}, map[string]any{
			"cpu_percent": vc[0],
		})
		if err != nil {
			logrus.Errorf("cat system info failed: %v", err)
		}

		vd, _ := disk.Usage("/")
		err = metric.LogPoint("system", "disk_usage", map[string]string{
			"hostName": hostInfo.Hostname,
		}, map[string]any{
			"disk_percent": vd.UsedPercent,
		})
		if err != nil {
			logrus.Errorf("cat system info failed: %v", err)
		}
	}
}
