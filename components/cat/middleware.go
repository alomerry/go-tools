package cat

import (
	"net/http"
	"strconv"
)

// HTTPMiddleware 返回通用 http.Handler 中间件：为每个请求创建事务、注入
// trace id、按响应码置状态并产出 transaction 点位。nameFn 返回事务名（建议
// 传路由模板函数，避免原始 path 的高基数）；为 nil 或返回空时回退 "METHOD path"。
// 响应体不做任何缓冲（仅包装捕获状态码），SSE 等流式响应不受影响。
func HTTPMiddleware(nameFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, ctx := NewTransactionWithCtx(r.Context(), transactionTypeHTTP, requestName(r, nameFn))
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			defer func() {
				if rec := recover(); rec != nil {
					tx.SetStatus(ERROR)
					tx.Complete()
					panic(rec)
				}
				if sw.status >= http.StatusBadRequest {
					tx.SetStatus(ERROR)
				}
				tx.AddData("status_code", strconv.Itoa(sw.status))
				tx.Complete()
			}()

			next.ServeHTTP(sw, r.WithContext(ctx))
		})
	}
}

func requestName(r *http.Request, nameFn func(*http.Request) string) string {
	var got string
	if nameFn != nil {
		got = nameFn(r)
	}
	return resolveName(got, r.Method+" "+r.URL.Path)
}

// statusWriter 仅捕获响应状态码，透传写操作与 Flusher，不缓冲响应体。
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if w.wrote {
		return
	}
	w.status = code
	w.wrote = true
	w.ResponseWriter.WriteHeader(code)
}

// Write 透传响应体并标记已写：net/http 语义下首次 Write 隐式发送 200，
// 此处同步置 wrote，保证后续 WriteHeader 被 guard 忽略、事务状态准确。
func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

// Flush 透传 http.Flusher，保证 SSE 流式推送可用。Flush 同样意味着响应头
// 已发送（未显式 WriteHeader 时按 200），故先置 wrote 再透传。
func (w *statusWriter) Flush() {
	if !w.wrote {
		w.wrote = true
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap 供 http.ResponseController 取底层 ResponseWriter。
func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
