package collect

import "time"

type ClientInfo struct {
	Hostname     string    `redis:"hostname"`
	LastReportAt time.Time `redis:"last_report_at"`
	CpuPercent   float64   `redis:"cpu_percent"`
	MemPercent   float64   `redis:"mem_percent"`
	DiskPercent  float64   `redis:"disk_percent"`
	NetBytesSent float64   `redis:"net_bytes_sent"`
	NetBytesRec  float64   `redis:"net_bytes_rec"`
	Version      string    `redis:"version"`
}
