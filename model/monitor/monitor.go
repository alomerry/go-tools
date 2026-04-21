package monitor

import "time"

type SystemStats struct {
	Timestamp time.Time

	Id          string
	Name        string
	PhysicalCPU int
	LogicalCPU  int
	CpuUsage    float64
	UserUsage   float64
	SystemUsage float64

	TotalMemory     uint64
	CachedMemory    uint64
	UsedMemory      uint64
	MaxMemory       uint64
	RssMemory       uint64
	PageFaults      uint64
	PageMajorFaults uint64
	MemoryUsage     float64

	DiskUsage map[string]float64
	LoadAvg   [3]float64

	Ip string
}
