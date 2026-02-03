package internal

import (
	"encoding/json"
	"fmt"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/kook/model"
)

// BaseService 基础服务结构
type BaseService struct {
	client http.Client
}

// NewBaseService 创建基础服务
func NewBaseService(client http.Client) *BaseService {
	return &BaseService{
		client: client,
	}
}

// execute 执行请求
func (s *BaseService) after(resp http.Response, result interface{}) error {
	var baseResp model.BaseResponse
	if err := json.Unmarshal(resp.Bytes(), &baseResp); err != nil {
		return fmt.Errorf("failed to unmarshal base response: %w", err)
	}
	// 检查响应状态码
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("query failed, %v %v", baseResp.Code, baseResp.Message)
	}

	// 解析结果
	if result == nil {
		return nil
	}

	raw, _ := json.Marshal(baseResp.Data)
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// if baseResp.Code != 0 || baseResp.Message != "ok"

	return nil
}
