package resty

import (
	"time"

	"github.com/alomerry/go-tools/components/cat"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/go-resty/resty/v2"
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
	tx := cat.TransactionFromCtx(resp.Request.Context())
	if tx == nil {
		return nil
	}

	if resp.StatusCode() >= 400 {
		tx.SetStatus(cat.ERROR)
		tx.AddData("status", resp.Status())
	} else {
		tx.SetStatus(cat.SUCCESS)
	}

	tx.Complete()
	return nil
}
