package sdk

import "time"

// SearchRequest 定义搜索请求参数
type SearchRequest struct {
	Index     string    `json:"index"`      // 索引名称
	Query     string    `json:"query"`      // 搜索关键词 (KQL 或 query_string)
	StartTime time.Time `json:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
	Page      int       `json:"page"`       // 页码 (从 1 开始)
	Size      int       `json:"size"`       // 每页大小
	Interval  string    `json:"interval"`   // 聚合时间间隔 (如 "1h", "1d")，为空则不聚合
}

// SearchResponse 定义搜索响应结构
type SearchResponse struct {
	Total     int64             `json:"total"`
	Hits      []interface{}     `json:"hits"`      // 原始日志数据
	Histogram []HistogramBucket `json:"histogram"` // 柱状图数据
}

// HistogramBucket 定义直方图聚合的一个桶
type HistogramBucket struct {
	Time  string `json:"time"`
	Count int64  `json:"count"`
}
