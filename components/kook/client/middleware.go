package client

import (
	"resty.dev/v3"
)

// processResponse 响应后处理
func (c *Client) processResponse(client *resty.Client, resp *resty.Response) error {
	// 处理速率限制头
	if rateLimit := resp.Header().Get("X-RateLimit-Limit"); rateLimit != "" {
		// c.updateRateLimit(rateLimit, resp.Header().Get("X-RateLimit-Remaining"))
	}

	// 处理分页链接
	if linkHeader := resp.Header().Get("Link"); linkHeader != "" {
		// 解析 Link 头，可用于后续分页请求
		// 可以存储到请求上下文或返回给调用者
	}

	return nil
}
