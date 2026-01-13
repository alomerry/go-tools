package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alomerry/go-tools/components/kook/model"
	"resty.dev/v3"
)

// BaseService 基础服务结构
type BaseService struct {
	client *resty.Client
}

// NewBaseService 创建基础服务
func NewBaseService(client *resty.Client) *BaseService {
	return &BaseService{
		client: client,
	}
}

// getRequest 获取请求实例
func (s *BaseService) getRequest(ctx context.Context) *resty.Request {
	req := s.client.R()

	// 设置上下文
	if ctx != nil {
		req.SetContext(ctx)
	}

	return req
}

// execute 执行请求
func (s *BaseService) execute(req *resty.Request, method, path string, result interface{}) (*resty.Response, error) {
	// 设置 URL 路径
	url := path

	// 执行请求
	var resp *resty.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = req.Get(url)
	case http.MethodPost:
		resp, err = req.Post(url)
	case http.MethodPut:
		resp, err = req.Put(url)
	case http.MethodPatch:
		resp, err = req.Patch(url)
	case http.MethodDelete:
		resp, err = req.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return resp, err
	}

	var baseResp model.BaseResponse
	if err := json.Unmarshal(resp.Bytes(), &baseResp); err != nil {
		return resp, fmt.Errorf("failed to unmarshal base response: %w", err)
	}
	// 检查响应状态码
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("query failed, %v %v", baseResp.Code, baseResp.Message)
	}

	// 解析结果
	if result == nil {
		return resp, nil
	}

	raw, _ := json.Marshal(baseResp.Data)
	if err := json.Unmarshal(raw, result); err != nil {
		return resp, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp, nil
}
