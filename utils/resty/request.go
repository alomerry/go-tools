package resty

import (
	"time"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils/trace"
	"resty.dev/v3"
)

func DefaultRequestMiddleware(client *resty.Client, req *resty.Request) error {
	// 从 context 获取额外参数
	if ctx := req.Context(); ctx != nil {
		// 处理上下文中的超时
		if deadline, ok := ctx.Deadline(); ok {
			client.SetTimeout(time.Until(deadline))
		}

		// 处理自定义 headers
		if headers, ok := ctx.Value(cons.CtxKeyHeaders).(map[string]string); ok {
			for k, v := range headers {
				req.SetHeader(k, v)
			}
		}
	}

	// 设置请求 ID 用于跟踪
	req.SetHeader(cons.TraceIdKey, trace.NewTraceId())

	return nil
}
