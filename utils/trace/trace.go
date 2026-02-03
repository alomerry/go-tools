package trace

import (
	"context"

	"github.com/alomerry/cat-go/message"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

func GetTraceId(ctx context.Context, defaultVal string) string {
	if ctx == nil {
		return defaultVal
	}

	tid := utils.FromCtx(ctx, cons.TraceIdKey)
	if len(tid) == 0 {
		tid = cast.ToString(ctx.Value(message.CtxKeyTransaction))
	}

	if len(tid) == 0 {
		return defaultVal
	}

	return tid
}

func GetOrNewTraceId(ctx context.Context) (string, bool) {
	tid := GetTraceId(ctx, "")
	if len(tid) == 0 {
		return uuid.New().String(), false
	}

	return tid, true
}

func GetOrNewTraceCtx(ctx context.Context) (context.Context, string) {
	tid, exists := GetOrNewTraceId(ctx)
	if ctx == nil {
		ctx = context.TODO()
	}

	if !exists {
		return context.WithValue(ctx, cons.TraceIdKey, tid), tid
	}
	return ctx, tid
}

func NewContext(ctx context.Context) context.Context {
	tid, exists := GetOrNewTraceId(ctx)
	if ctx == nil {
		ctx = context.TODO()
	}

	if !exists {
		return context.WithValue(ctx, cons.TraceIdKey, tid)
	}
	return ctx
}
