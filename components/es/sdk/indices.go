package sdk

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetIndices 获取所有索引列表
func (s *Service) GetIndices(ctx context.Context) ([]string, error) {
	resp, err := s.client.Cat.Indices().Format("json").Do(ctx)
	if err != nil {
		return nil, err
	}

	var indices []string
	for _, record := range resp {
		if record.Index != nil {
			indices = append(indices, *record.Index)
		}
	}
	return indices, nil
}

// GetMapping 获取指定索引的字段映射
func (s *Service) GetMapping(ctx context.Context, index string) (map[string]interface{}, error) {
	resp, err := s.client.Indices.GetMapping().Index(index).Do(ctx)
	if err != nil {
		return nil, err
	}

	if record, ok := resp[index]; ok {
		// Convert to map
		b, err := json.Marshal(record.Mappings)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, fmt.Errorf("index %s not found", index)
}
