package utils

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func FromCtx(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// grpc metadata 的 key 规范为小写，按原始大小写查不到
		if val, ok := md[strings.ToLower(key)]; ok {
			return strings.Join(val, ",")
		}

		return ""
	}

	// 兼容既有调用方以 string 作为 ctx key 的用法（非 Go 惯例，
	// 存在 key 冲突风险，但变更会断既有数据链，保持原状并标注）
	val, ok := ctx.Value(key).(string)
	if ok {
		return val
	}

	return ""
}
