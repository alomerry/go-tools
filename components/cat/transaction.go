package cat

import (
	"context"
	"sync"
	"time"

	"github.com/alomerry/go-tools/static/cons"
)

// Transactor 表示一个 CAT 风格的事务（调用段打点）。
type Transactor interface {
	// Complete 结束事务并产出 transaction 点位，重复调用只计一次。
	Complete()
	// SetStatus 设置事务状态（SUCCESS / ERROR 或自定义字符串）。
	SetStatus(status string)
	// AddData 附加键值对，作为 transaction 点位的 string field 落库。
	AddData(k, v string)
	// TraceId 返回事务的 trace id（创建时继承或生成）。
	TraceId() string
}

type transaction struct {
	mu       sync.Mutex
	mtype    string
	name     string
	status   string
	data     map[string]string
	start    time.Time
	traceId  string
	complete bool
}

// NewTransactionWithCtx 创建事务并返回携带事务的 ctx（同时注入 cons.TraceIdKey，
// 供日志组件与下游中间件提取 trace id）。ctx 已有 trace id 时继承；
// 未初始化时返回空实现，ctx 注入行为不变。
func NewTransactionWithCtx(ctx context.Context, mtype, name string) (Transactor, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	tid := newTraceId(ctx)
	nctx := context.WithValue(ctx, cons.TraceIdKey, tid)

	if !isEnabled() {
		return nullTransaction{traceId: tid}, nctx
	}

	t := &transaction{
		mtype:   mtype,
		name:    name,
		status:  SUCCESS,
		data:    make(map[string]string),
		start:   time.Now(),
		traceId: tid,
	}
	return t, context.WithValue(nctx, ctxKeyTransaction{}, t)
}

func (t *transaction) TraceId() string {
	return t.traceId
}

func (t *transaction) SetStatus(status string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if status != "" {
		t.status = status
	}
}

func (t *transaction) AddData(k, v string) {
	if k == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.data) >= maxTransactionData {
		return
	}
	t.data[k] = truncateString(v, maxDataValue)
}

func (t *transaction) Complete() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.complete {
		return
	}
	t.complete = true

	fields := make(map[string]any, len(t.data)+1)
	fields["duration_ms"] = time.Since(t.start).Milliseconds()
	for k, v := range t.data {
		fields[k] = v
	}

	emit(point{
		measurement: "transaction",
		tags: map[string]string{
			"service": serviceName(),
			"type":    t.mtype,
			"name":    t.name,
			"status":  t.status,
		},
		fields: fields,
	})
}

// nullTransaction 未初始化时的空实现：保留 trace id 供日志链路使用，不做任何写入。
type nullTransaction struct {
	traceId string
}

func (nullTransaction) Complete()              {}
func (nullTransaction) SetStatus(string)       {}
func (nullTransaction) AddData(string, string) {}
func (t nullTransaction) TraceId() string      { return t.traceId }

// TransactionFromCtx 取出 ctx 中的当前事务；不存在时返回 nil。
// 供中间件响应阶段直接操作事务对象（如 resty）；日志 hook 等单值场景用 AddData。
func TransactionFromCtx(ctx context.Context) Transactor {
	if ctx == nil {
		return nil
	}
	t, _ := ctx.Value(ctxKeyTransaction{}).(Transactor)
	return t
}

// AddData 向当前事务附加一条日志数据，固定以 "log" 为 field 名。
// 仅供日志 hook 等无键调用方使用，其余场景直接用事务对象的 AddData(k, v)。
func AddData(ctx context.Context, value string) {
	if t := TransactionFromCtx(ctx); t != nil {
		t.AddData("log", value)
	}
}
