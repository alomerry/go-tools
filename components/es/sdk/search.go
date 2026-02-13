package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"time"
)

// SearchLogs 执行日志搜索
func (s *Service) SearchLogs(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	// 1. 构建 JSON 请求体
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"filter": []interface{}{
				map[string]interface{}{
					"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gte": req.StartTime.Format(time.RFC3339),
							"lte": req.EndTime.Format(time.RFC3339),
						},
					},
				},
			},
		},
	}

	if req.Query != "" {
		boolQ := query["bool"].(map[string]interface{})
		boolQ["must"] = []interface{}{
			map[string]interface{}{
				"query_string": map[string]interface{}{
					"query": req.Query,
				},
			},
		}
	}

	from := (req.Page - 1) * req.Size
	if from < 0 {
		from = 0
	}

	bodyMap := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  req.Size,
		"sort": []interface{}{
			map[string]interface{}{
				"@timestamp": "desc",
			},
		},
	}

	if req.Interval != "" {
		bodyMap["aggs"] = map[string]interface{}{
			"histogram": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "@timestamp",
					"fixed_interval": req.Interval,
					"min_doc_count":  0,
				},
			},
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(bodyMap); err != nil {
		return nil, err
	}

	// 2. 执行查询 (使用 Raw 方式避免复杂的类型构建)
	resp, err := s.client.Search().Index(req.Index).Raw(&buf).Do(ctx)
	if err != nil {
		return nil, err
	}

	// 3. 解析响应
	searchResp := &SearchResponse{
		Hits:      make([]interface{}, 0),
		Histogram: make([]HistogramBucket, 0),
	}

	if resp.Hits.Total != nil {
		searchResp.Total = resp.Hits.Total.Value
	}
	for _, hit := range resp.Hits.Hits {
		var source interface{}
		if err := json.Unmarshal(hit.Source_, &source); err == nil {
			searchResp.Hits = append(searchResp.Hits, source)
		}
	}

	if resp.Aggregations != nil {
		if agg, ok := resp.Aggregations["histogram"]; ok {
			// 使用 JSON 中转来解析聚合结果，因为 Aggregate 类型较复杂
			b, _ := json.Marshal(agg)
			
			// 尝试解析 DateHistogram 结构
			var dhAgg struct {
				DateHistogram struct {
					Buckets []struct {
						KeyAsString string `json:"key_as_string"`
						DocCount    int64  `json:"doc_count"`
						Key         int64  `json:"key"`
					} `json:"buckets"`
				} `json:"date_histogram"`
			}
			
			// 注意：TypedClient 的 Aggregations 序列化后可能直接就是聚合内容，或者包含类型包装
			// 我们先尝试解析包含 date_histogram 包装的情况
			if err := json.Unmarshal(b, &dhAgg); err == nil && len(dhAgg.DateHistogram.Buckets) > 0 {
				for _, bucket := range dhAgg.DateHistogram.Buckets {
					timeStr := bucket.KeyAsString
					if timeStr == "" {
						timeStr = time.UnixMilli(bucket.Key).Format(time.RFC3339)
					}
					searchResp.Histogram = append(searchResp.Histogram, HistogramBucket{
						Time:  timeStr,
						Count: bucket.DocCount,
					})
				}
			} else {
				// 尝试直接解析 buckets (如果 agg 直接是 DateHistogramAggregate)
				var bucketsAgg struct {
					Buckets []struct {
						KeyAsString string `json:"key_as_string"`
						DocCount    int64  `json:"doc_count"`
						Key         int64  `json:"key"`
					} `json:"buckets"`
				}
				if err := json.Unmarshal(b, &bucketsAgg); err == nil {
					for _, bucket := range bucketsAgg.Buckets {
						timeStr := bucket.KeyAsString
						if timeStr == "" {
							timeStr = time.UnixMilli(bucket.Key).Format(time.RFC3339)
						}
						searchResp.Histogram = append(searchResp.Histogram, HistogramBucket{
							Time:  timeStr,
							Count: bucket.DocCount,
						})
					}
				}
			}
		}
	}

	return searchResp, nil
}
