package cat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/cons"
)

// capturePoints 临时替换 emit 捕获点位，返回点位 channel。
func capturePoints(t *testing.T) <-chan point {
	t.Helper()
	ch := make(chan point, 16)
	orig := emit
	emit = func(p point) { ch <- p }
	t.Cleanup(func() { emit = orig })
	return ch
}

// resetCat 恢复未初始化状态，避免用例间串扰。
func resetCat(t *testing.T) {
	t.Helper()
	setCatState(false, "")
	t.Cleanup(func() { setCatState(false, "") })
}

func setCatState(enabledVal bool, svcVal string) {
	mu.Lock()
	defer mu.Unlock()
	enabled = enabledVal
	svc = svcVal
}

func drainPoint(t *testing.T, ch <-chan point) point {
	t.Helper()
	select {
	case p := <-ch:
		return p
	case <-time.After(time.Second):
		t.Fatal("expected a point to be emitted")
		return point{}
	}
}

func assertNoPoint(t *testing.T, ch <-chan point) {
	t.Helper()
	select {
	case p := <-ch:
		t.Fatalf("unexpected point emitted: %+v", p)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestTransactionLifecycle(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	tx, ctx := NewTransactionWithCtx(context.Background(), "RPC", "/backend.svc.Svc/Method")
	if tx == nil {
		t.Fatal("transaction should not be nil")
	}
	tid := tx.TraceId()
	if tid == "" {
		t.Fatal("trace id should not be empty")
	}
	if got := ctx.Value(cons.TraceIdKey); got != tid {
		t.Fatalf("ctx trace id = %v, want %v", got, tid)
	}
	if TraceIdFromCtx(ctx) != tid {
		t.Fatalf("TraceIdFromCtx = %q, want %q", TraceIdFromCtx(ctx), tid)
	}

	tx.AddData("error", "boom")
	tx.SetStatus(ERROR)
	tx.Complete()
	tx.Complete() // 重复 Complete 只计一次

	p := drainPoint(t, ch)
	assertNoPoint(t, ch)

	if p.measurement != "transaction" {
		t.Fatalf("measurement = %q, want transaction", p.measurement)
	}
	wantTags := map[string]string{
		"service": "svc-a",
		"type":    "RPC",
		"name":    "/backend.svc.Svc/Method",
		"status":  ERROR,
	}
	if !reflect.DeepEqual(p.tags, wantTags) {
		t.Fatalf("tags = %+v, want %+v", p.tags, wantTags)
	}
	duration, ok := p.fields["duration_ms"].(int64)
	if !ok {
		t.Fatalf("duration_ms field missing or wrong type: %T", p.fields["duration_ms"])
	}
	if duration < 0 {
		t.Fatalf("duration_ms = %d, want >= 0", duration)
	}
	if p.fields["error"] != "boom" {
		t.Fatalf("data field error = %v, want boom", p.fields["error"])
	}

	// 未显式 SetStatus 的事务默认 SUCCESS
	tx2, _ := NewTransactionWithCtx(context.Background(), "URL", "GET /ok")
	tx2.Complete()

	p2 := drainPoint(t, ch)
	if p2.tags["status"] != SUCCESS {
		t.Fatalf("default status = %q, want %q", p2.tags["status"], SUCCESS)
	}
}

func TestTransactionInheritsParentTraceId(t *testing.T) {
	resetCat(t)
	Init("svc-a")

	parent, pctx := NewTransactionWithCtx(context.Background(), "URL", "GET /a")
	_, cctx := NewTransactionWithCtx(pctx, "RPC", "/backend.svc.Svc/Method")

	if got := cctx.Value(cons.TraceIdKey); got != parent.TraceId() {
		t.Fatalf("child trace id = %v, want inherit %v", got, parent.TraceId())
	}
}

func TestDisabledNoopButTraceIdInjected(t *testing.T) {
	resetCat(t) // 未 Init
	ch := capturePoints(t)

	tx, ctx := NewTransactionWithCtx(context.Background(), "URL", "GET /x")
	if tx == nil {
		t.Fatal("null transaction should not be nil")
	}
	if tx.TraceId() == "" {
		t.Fatal("null transaction should still carry a trace id")
	}
	if ctx.Value(cons.TraceIdKey) == "" {
		t.Fatal("trace id should still be injected into ctx when disabled")
	}

	tx.SetStatus(ERROR)
	tx.AddData("k", "v")
	tx.Complete()
	assertNoPoint(t, ch)

	// ctx 便捷函数在无事务/禁用时安全空操作
	if tx := TransactionFromCtx(ctx); tx != nil {
		tx.SetStatus(ERROR)
		tx.AddData("k", "v")
		tx.Complete()
	}
	AddData(ctx, "v")
	if TransactionFromCtx(nil) != nil {
		t.Fatal("TransactionFromCtx(nil) should be nil")
	}
	assertNoPoint(t, ch)

	// NewTransactionWithCtx(nil, ...) 不 panic
	tx2, _ := NewTransactionWithCtx(nil, "URL", "GET /nil")
	tx2.Complete()
	assertNoPoint(t, ch)

	LogEvent(ctx, "E", "n")
	assertNoPoint(t, ch)

	LogError(ctx, errors.New("x"))
	assertNoPoint(t, ch)
}

func TestLogEvent(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	LogEvent(context.Background(), "AgentClient", "10.0.0.1:offline")
	p := drainPoint(t, ch)

	if p.measurement != "event" {
		t.Fatalf("measurement = %q, want event", p.measurement)
	}
	wantTags := map[string]string{
		"service": "svc-a",
		"type":    "AgentClient",
		"name":    "10.0.0.1:offline",
	}
	if !reflect.DeepEqual(p.tags, wantTags) {
		t.Fatalf("tags = %+v, want %+v", p.tags, wantTags)
	}
	if p.fields["data"] != "" {
		t.Fatalf("data = %v, want empty", p.fields["data"])
	}

	// args 并入 data
	LogEvent(context.Background(), "E", "n", "k=v", "plain")
	p2 := drainPoint(t, ch)
	if data := p2.fields["data"].(string); data != "k=v\nplain" {
		t.Fatalf("data = %q, want %q", data, "k=v\nplain")
	}
}

func TestLogError(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	// fmt.Errorf（无 %w）产物为 errorString，无语义推导不出类别 → 兜底
	LogError(context.Background(), fmt.Errorf("db timeout"))
	p := drainPoint(t, ch)

	if p.measurement != "problem" {
		t.Fatalf("measurement = %q, want problem", p.measurement)
	}
	if p.tags["service"] != "svc-a" {
		t.Fatalf("service tag = %q, want svc-a", p.tags["service"])
	}
	if p.tags["type"] != fallbackProblemType {
		t.Fatalf("type tag = %q, want fallback %q", p.tags["type"], fallbackProblemType)
	}
	if name, ok := p.tags["name"]; ok {
		t.Fatalf("problem should not carry name tag, got %q", name)
	}
	msg := p.fields["message"].(string)
	if msg != "db timeout" {
		t.Fatalf("message = %q, want single original text", msg)
	}

	// 语义化错误类型推导出类别，且不依赖调用方栈帧
	LogError(context.Background(), &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist})
	p2 := drainPoint(t, ch)
	if got := p2.tags["type"]; got != "PathError" {
		t.Fatalf("type tag = %q, want PathError", got)
	}

	// args 并入 message，空段跳过
	LogError(context.Background(), errors.New("raw"), "extra ctx", "")
	p3 := drainPoint(t, ch)
	msg3 := p3.fields["message"].(string)
	if msg3 != "raw · extra ctx" {
		t.Fatalf("message = %q, want %q", msg3, "raw · extra ctx")
	}

	// nil error 安全
	LogError(context.Background(), nil)
	assertNoPoint(t, ch)
}

// TestLogErrorWithType 覆盖显式 type 口径（logHook 路径）：typ 直写 tag、
// 空 typ 落 fallbackProblemType、超长 typ 截断、extras 并入 message；并验证
// LogError 经 err 推导委托后行为不变。
func TestLogErrorWithType(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	// 显式 typ（Errorf 格式串形态）直写 tag，extras 并入 message
	LogErrorWithType(context.Background(), "db query failed: %v", errors.New("db query failed: timeout"), "k=v")
	p := drainPoint(t, ch)
	if p.measurement != "problem" {
		t.Fatalf("measurement = %q, want problem", p.measurement)
	}
	if p.tags["type"] != "db query failed: %v" {
		t.Fatalf("type tag = %q, want format string", p.tags["type"])
	}
	if msg := p.fields["message"].(string); msg != "db query failed: timeout · k=v" {
		t.Fatalf("message = %q, want extras appended", msg)
	}

	// 空 typ（Error 无格式串形态且文本为空）落兜底
	LogErrorWithType(context.Background(), "", errors.New("x"))
	if p2 := drainPoint(t, ch); p2.tags["type"] != fallbackProblemType {
		t.Fatalf("type tag = %q, want fallback %q", p2.tags["type"], fallbackProblemType)
	}

	// 超长 typ 截断到 maxProblemType
	LogErrorWithType(context.Background(), strings.Repeat("t", maxProblemType+100), errors.New("x"))
	if p3 := drainPoint(t, ch); len(p3.tags["type"]) != maxProblemType {
		t.Fatalf("type tag len = %d, want %d", len(p3.tags["type"]), maxProblemType)
	}

	// 换行归一为空格：type tag 保持单行卫生
	LogErrorWithType(context.Background(), "line1\nline2\rline3", errors.New("x"))
	if p := drainPoint(t, ch); p.tags["type"] != "line1 line2 line3" {
		t.Fatalf("type tag = %q, want newlines normalized", p.tags["type"])
	}

	// nil error 安全
	LogErrorWithType(context.Background(), "typ", nil)
	assertNoPoint(t, ch)

	// LogError 直打路径（err 推导委托）行为不变：语义类型仍推导、通用类型仍兜底
	LogError(context.Background(), &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist})
	if p4 := drainPoint(t, ch); p4.tags["type"] != "PathError" {
		t.Fatalf("type tag = %q, want PathError", p4.tags["type"])
	}
	LogError(context.Background(), errors.New("boom"))
	if p5 := drainPoint(t, ch); p5.tags["type"] != fallbackProblemType {
		t.Fatalf("type tag = %q, want fallback %q", p5.tags["type"], fallbackProblemType)
	}
}

// TestErrorCategory 表驱动覆盖错误类别推导：具体类型名提取（去包名/指针）、
// 解包链取首个具语义类型名、通用无语义类型兜底、panic 运行时错误的自然推导。
func TestErrorCategory(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"path-error", &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist}, "PathError"},
		{"error-string-no-semantics", errors.New("boom"), ""},
		{"wrap-error-unwraps", fmt.Errorf("call failed: %w", &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist}), "PathError"},
		{"nested-wrap-unwraps", fmt.Errorf("a: %w", fmt.Errorf("b: %w", &fs.PathError{Op: "open", Path: "/x", Err: fs.ErrNotExist})), "PathError"},
		{"wrapped-generic-falls-back", fmt.Errorf("a: %w", errors.New("inner")), ""},
		{"multi-wrapped-generics-fall-back", fmt.Errorf("a %w b %w", errors.New("x"), errors.New("y")), ""},
		{"wrapped-join-falls-back", fmt.Errorf("a: %w", errors.Join(errors.New("x"), errors.New("y"))), ""},
		{"join-with-semantic-falls-back", errors.Join(&cronJobError{msg: "job failed"}, errors.New("x")), ""},
		{"custom-type", &cronJobError{msg: "job failed"}, "cronJobError"},
		{"wrapped-custom-type", fmt.Errorf("ctx: %w", &cronJobError{msg: "job failed"}), "cronJobError"},
		{"runtime-bounds-panic", recoveredBoundsPanic(), "boundsError"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := errorCategory(c.err); got != c.want {
				t.Fatalf("errorCategory(%T) = %q, want %q", c.err, got, c.want)
			}
		})
	}
}

// cronJobError 测试用自定义错误类型，模拟业务语义化错误。
type cronJobError struct {
	msg string
}

func (e *cronJobError) Error() string { return e.msg }

// recoveredBoundsPanic 触发切片越界 panic 并恢复为 error，模拟运行时 panic 错误
// 经 LogError 打点的来源形态（defer 恢复 + error 接口捕获）。
func recoveredBoundsPanic() (err error) {
	defer func() { err, _ = recover().(error) }()
	var s []int
	_ = s[1] //nolint:govet // 刻意越界触发 runtime.boundsError
	return
}

func TestJoinMessage(t *testing.T) {
	cases := []struct {
		name  string
		first string
		rest  []string
		want  string
	}{
		{"single", "a", nil, "a"},
		{"join", "a", []string{"k=v", "b"}, "a · k=v · b"},
		{"skip-empty", "a", []string{"", "b"}, "a · b"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := joinMessage(c.first, c.rest...); got != c.want {
				t.Fatalf("joinMessage(%q, %v) = %q, want %q", c.first, c.rest, got, c.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"no-trunc", "abc", 10, "abc"},
		{"ascii", "abcdef", 3, "abc"},
		{"exact", "abcd", 4, "abcd"},
		{"zero", "abc", 0, ""},
		{"negative", "abc", -1, ""},
		{"multibyte-boundary", "你好", 4, "你"},
		{"multibyte-to-empty", "你好", 2, ""},
		{"multibyte-exact", "你好", 6, "你好"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := truncateString(c.in, c.max)
			if got != c.want {
				t.Fatalf("truncateString(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
			}
			if got != "" && !strings.HasPrefix(c.in, got) {
				t.Fatalf("result %q is not a prefix of %q", got, c.in)
			}
		})
	}
}

// TestEmitLimits 表驱动覆盖三类点位长度上限：event data、problem message、
// transaction data 条数。纯函数边界见 TestTruncateString。
func TestEmitLimits(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	t.Run("event data truncated", func(t *testing.T) {
		LogEvent(context.Background(), "E", "n", strings.Repeat("x", maxEventData+100))
		if got := len(drainPoint(t, ch).fields["data"].(string)); got != maxEventData {
			t.Fatalf("data len = %d, want %d", got, maxEventData)
		}

		LogEvent(context.Background(), "E", "n", strings.Repeat("y", 1500), strings.Repeat("z", 1500))
		if got := len(drainPoint(t, ch).fields["data"].(string)); got != maxEventData {
			t.Fatalf("joined data len = %d, want %d", got, maxEventData)
		}
	})

	t.Run("problem message truncated", func(t *testing.T) {
		LogError(context.Background(), errors.New(strings.Repeat("y", 4096)))
		if got := len(drainPoint(t, ch).fields["message"].(string)); got != maxProblemMessage {
			t.Fatalf("message len = %d, want %d", got, maxProblemMessage)
		}
	})

	t.Run("transaction data capped", func(t *testing.T) {
		tx, _ := NewTransactionWithCtx(context.Background(), "URL", "GET /cap")
		for i := 0; i < 100; i++ {
			tx.AddData(fmt.Sprintf("k%d", i), "v")
		}
		tx.Complete()

		if got := len(drainPoint(t, ch).fields); got != maxTransactionData+1 { // + duration_ms
			t.Fatalf("fields count = %d, want %d", got, maxTransactionData+1)
		}
	})
}

// TestPointOptions 轻量断言点位到 tsdb options 的接线：metric + tags +
// fields 数量对齐，field 经 WithFieldAny 直写（string 不进 Tags 由 tsdb 侧保证）。
func TestPointOptions(t *testing.T) {
	p := point{
		measurement: "transaction",
		tags: map[string]string{
			"service": "s", "type": "URL", "name": "n", "status": SUCCESS,
		},
		fields: map[string]any{
			"duration_ms": int64(12),
			"data":        "中文字符串 field",
		},
	}

	opts := p.options()
	if want := 1 + len(p.tags) + len(p.fields); len(opts) != want {
		t.Fatalf("options count = %d, want %d", len(opts), want)
	}
}

func TestHTTPMiddlewareStatusAndName(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	nameFn := func(r *http.Request) string { return "GET /api/items/{id}" }
	handler := HTTPMiddleware(nameFn)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(cons.TraceIdKey) == "" {
			t.Error("trace id missing in request ctx")
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/items/123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}

	p := drainPoint(t, ch)
	if p.tags["name"] != "GET /api/items/{id}" {
		t.Fatalf("name = %q, want route template", p.tags["name"])
	}
	if p.tags["status"] != ERROR {
		t.Fatalf("status = %q, want %q for 404", p.tags["status"], ERROR)
	}
	if p.fields["status_code"] != "404" {
		t.Fatalf("status_code = %v, want 404", p.fields["status_code"])
	}
}

func TestHTTPMiddlewareDefaultNameAndSuccess(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	t.Run("nil nameFn falls back to METHOD path", func(t *testing.T) {
		handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok")) // 隐式 200
		}))

		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		p := drainPoint(t, ch)
		if p.tags["name"] != "POST /x" {
			t.Fatalf("name = %q, want 'POST /x'", p.tags["name"])
		}
		if p.tags["status"] != SUCCESS {
			t.Fatalf("status = %q, want %q", p.tags["status"], SUCCESS)
		}
	})

	t.Run("empty nameFn result falls back to METHOD path", func(t *testing.T) {
		handler := HTTPMiddleware(func(r *http.Request) string { return "" })(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		p := drainPoint(t, ch)
		if p.tags["name"] != "GET /fallback" {
			t.Fatalf("name = %q, want 'GET /fallback'", p.tags["name"])
		}
	})
}

// TestHTTPMiddlewareStreams 验证中间件不缓冲响应体：handler Flush 后客户端
// 必须能在 handler 结束前读到首块（SSE 场景）。若中间件缓冲 body，
// 首块读取会阻塞到 handler 退出（而 handler 又在等测试放行）→ 超时判失败。
func TestHTTPMiddlewareStreams(t *testing.T) {
	resetCat(t)
	Init("svc-a")

	flushed := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	signal := func() { releaseOnce.Do(func() { close(release) }) }
	defer signal() // 超时分支也要放行 handler，避免 httptest.Server.Close 卡住

	ts := httptest.NewServer(HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("chunk-1"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(flushed)
		<-release // 等「首块已被客户端读到」再继续，若被缓冲这里会死锁
		_, _ = w.Write([]byte("chunk-2"))
	})))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	first := make(chan string, 1)
	go func() {
		b := make([]byte, len("chunk-1"))
		if _, err := io.ReadFull(resp.Body, b); err != nil {
			first <- "read error: " + err.Error()
			return
		}
		first <- string(b)
	}()

	select {
	case got := <-first:
		if got != "chunk-1" {
			t.Fatalf("first chunk = %q, want chunk-1", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("middleware buffered the response body: first chunk not streamed while handler still running")
	}

	signal() // 放行 handler 写 chunk-2
	rest, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read rest failed: %v", err)
	}
	if string(rest) != "chunk-2" {
		t.Fatalf("rest = %q, want chunk-2", rest)
	}
}

func TestHTTPMiddlewarePanic(t *testing.T) {
	resetCat(t)
	Init("svc-a")
	ch := capturePoints(t)

	handler := HTTPMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	func() {
		defer func() { _ = recover() }()
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/p", nil))
	}()

	p := drainPoint(t, ch)
	if p.tags["status"] != ERROR {
		t.Fatalf("status = %q, want %q after panic", p.tags["status"], ERROR)
	}
}
