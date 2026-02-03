package resty

import (
	"time"

	"github.com/alomerry/cat-go/cat"
	"github.com/alomerry/go-tools/static/cons"
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

	_, ctx := cat.NewTransactionWithCtx(req.Context(), "URL", req.URL)
	req.SetContext(ctx)

	return nil
}

func DefaultResponseMiddleware(client *resty.Client, resp *resty.Response) error {
	ctx := resp.Request.Context()

	if resp.StatusCode() >= 400 {
		cat.SetStatus(ctx, cat.ERROR)
		cat.AddData(ctx, "status", resp.Status())
	} else {
		cat.SetStatus(ctx, cat.SUCCESS)
	}

	cat.CompleteTransaction(ctx)
	return nil
}
