package monitor

import "time"

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

	Ip string
}
