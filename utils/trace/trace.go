package trace

import (
	"context"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/google/uuid"
)

func NewTraceId() string {
	return uuid.New().String()
}

func NewCtx(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}
	return context.WithValue(ctx, cons.TraceIdKey, NewTraceId())
}
