package utils

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func FromCtx(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if val, ok := md[key]; ok {
			return strings.Join(val, ",")
		}

		return ""
	}

	val, ok := ctx.Value(key).(string)
	if ok {
		return val
	}

	return ""
}
