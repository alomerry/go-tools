package log

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// captureHook 捕获 logrus entry，断言 builder 链的 Message/Data 口径。
type captureHook struct {
	mu      sync.Mutex
	entries []*logrus.Entry
}

func (h *captureHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *captureHook) Fire(e *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	cp := *e
	cp.Data = make(logrus.Fields, len(e.Data))
	for k, v := range e.Data {
		cp.Data[k] = v
	}
	h.entries = append(h.entries, &cp)
	return nil
}

func (h *captureHook) last() *logrus.Entry {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.entries) == 0 {
		return nil
	}
	return h.entries[len(h.entries)-1]
}

func installCaptureHook(t *testing.T) *captureHook {
	t.Helper()
	std := logrus.StandardLogger()
	origHooks := make(logrus.LevelHooks, len(std.Hooks))
	for lvl, hooks := range std.Hooks {
		origHooks[lvl] = append([]logrus.Hook(nil), hooks...)
	}
	origOut := std.Out
	origLevel := std.GetLevel()
	t.Cleanup(func() {
		std.Hooks, std.Out = origHooks, origOut
		std.SetLevel(origLevel)
	})
	h := &captureHook{}
	std.Hooks = make(logrus.LevelHooks)
	std.AddHook(h)
	std.SetOutput(io.Discard)
	std.SetLevel(logrus.DebugLevel)
	return h
}

// TestBuilderFirstStableMessage 验证目标调用形态：先 With 后 Errorf，
// format 为稳定常量时 entry.Message 即稳定常量，可变细节全进 Data。
func TestBuilderFirstStableMessage(t *testing.T) {
	h := installCaptureHook(t)
	ctx := context.Background()
	err := errors.New("dial refused")

	WithError(err).WithField("method", "GET").WithField("path", "/ping1").Errorf(ctx, "gateway upstream unavailable")
	e := h.last()
	if e == nil {
		t.Fatal("expected an entry to be captured")
	}
	if e.Message != "gateway upstream unavailable" {
		t.Fatalf("message = %q, want stable constant", e.Message)
	}
	if e.Data["method"] != "GET" || e.Data["path"] != "/ping1" {
		t.Fatalf("data missing method/path: %+v", e.Data)
	}
	if got, ok := e.Data[logrus.ErrorKey]; !ok || got != err {
		t.Fatalf("data[%q] = %v (%v), want original err", logrus.ErrorKey, got, ok)
	}

	// WithFields 批量形态同样透传。
	WithFields(logrus.Fields{"a": 1}).WithError(err).Errorf(ctx, "stable")
	e2 := h.last()
	if e2.Message != "stable" || e2.Data["a"] != 1 {
		t.Fatalf("WithFields chain lost fields: %+v", e2.Data)
	}
}

// TestWithErrorNilSafe 验证 nil err 不注入 error 字段、不污染 extras。
func TestWithErrorNilSafe(t *testing.T) {
	h := installCaptureHook(t)
	ctx := context.Background()

	WithError(nil).WithField("path", "/x").Errorf(ctx, "stable")
	if e := h.last(); e == nil {
		t.Fatal("expected an entry")
	} else if _, ok := e.Data[logrus.ErrorKey]; ok {
		t.Fatalf("nil err must not inject %q field: %+v", logrus.ErrorKey, e.Data)
	}
}

// TestLegacyCallStillFormats 验证老调用零改动可编译且行为 fallback 不变：
// format 变量仍按 logrus 语义展开为动态文本（hook 侧 type 取该文本，现状口径）。
func TestLegacyCallStillFormats(t *testing.T) {
	h := installCaptureHook(t)
	ctx := context.Background()

	Errorf(ctx, "not found: %s %s", "GET", "/ping1")
	if e := h.last(); e == nil || e.Message != "not found: GET /ping1" {
		t.Fatalf("legacy Errorf must expand format, got %+v", e)
	}
}
