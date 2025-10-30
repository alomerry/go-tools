package grpc

import (
	"github.com/alomerry/go-tools/static/cons"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/introduction/

func CustomMatcher(key string) (string, bool) {
	switch key {
	case cons.TraceIdKey:
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
