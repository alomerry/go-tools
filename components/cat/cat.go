package cat

import (
	"context"
	"sync"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/rs/xid"
)

// 事务状态常量。取可读值，便于 Grafana 侧按 status 过滤与告警。
const (
	SUCCESS = "success"
	ERROR   = "error"
)

// 事务/问题点位类型，对齐 CAT 的事务类型命名。problem 点位的 type 为调用方
// 包路径（见 problem.go），故此处仅保留事务类型。
const (
	transactionTypeHTTP = "URL"
	transactionTypeRPC  = "RPC"
)

var (
	mu      sync.RWMutex
	enabled bool
	svc     string
)

// Init 启用 cat 打点，service 为点位上的 service tag（一般传 env.GetService()）。
// 未调用 Init 前所有 API 走空操作：不写入、不阻塞、不 panic，
// 但 trace id 仍会注入 ctx，保证日志链路的 trace id 不断链。
func Init(service string) {
	mu.Lock()
	defer mu.Unlock()
	enabled = true
	if service != "" {
		svc = service
	}
}

func isEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return enabled
}

func serviceName() string {
	mu.RLock()
	defer mu.RUnlock()
	return svc
}

// ctxKeyTransaction 私有 ctx key，避免与其他组件的 string key 冲突。
type ctxKeyTransaction struct{}

// resolveName 归一化事务名：调用方归一化结果非空则采用，否则回退 fallback。
func resolveName(got, fallback string) string {
	if got != "" {
		return got
	}
	return fallback
}

// TraceIdFromCtx 返回 ctx 中当前事务的 TraceId；无事务时返回空串。
func TraceIdFromCtx(ctx context.Context) string {
	if t := TransactionFromCtx(ctx); t != nil {
		return t.TraceId()
	}
	return ""
}

// newTraceId 继承或生成 trace id：ctx 已带 trace id 时继承，否则新生成。
// （事务创建时必然同步注入了等值 cons.TraceIdKey，故无需再查事务对象。）
func newTraceId(ctx context.Context) string {
	if id, ok := ctx.Value(cons.TraceIdKey).(string); ok && id != "" {
		return id
	}
	return xid.New().String()
}
