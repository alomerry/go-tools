package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// captureHook 捕获流经 logrus 的 entry，用于断言注入行为。
type captureHook struct {
	entries []*logrus.Entry
}

func (h *captureHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *captureHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)
	return nil
}

// installCaptureHook 注册捕获 hook 并保存 logrus 全局状态，测试结束后恢复，
// 避免用例间（及与 logrus_format_test.go 间）串扰。Hooks 须深拷贝：AddHook
// 原地修改 map，浅存引用则恢复形同虚设。
func installCaptureHook(t *testing.T) *captureHook {
	t.Helper()
	std := logrus.StandardLogger()
	origHooks := make(logrus.LevelHooks, len(std.Hooks))
	for lvl, hooks := range std.Hooks {
		origHooks[lvl] = append([]logrus.Hook(nil), hooks...)
	}
	origFormatter, origOut, origReportCaller := std.Formatter, std.Out, std.ReportCaller
	t.Cleanup(func() {
		std.Hooks, std.Formatter, std.Out, std.ReportCaller = origHooks, origFormatter, origOut, origReportCaller
	})
	h := &captureHook{}
	logrus.AddHook(h)
	return h
}

// TestErrorfInjectsReservedProblemTypeField 覆盖注入口径：仅 Errorf（含 Logger
// 方法版）注入保留字段且不影响 message 本体；其他级别与 Error 的无格式串形态
// 不注入（由 hook 侧动态兜底）。
func TestErrorfInjectsReservedProblemTypeField(t *testing.T) {
	cases := []struct {
		name      string
		log       func(h *captureHook)
		wantType  string // 保留字段期望值；空串表示不注入
		wantExtra bool   // WithField 附加字段是否应保留
	}{
		{"errorf-injects", func(h *captureHook) {
			Errorf(context.Background(), "boom %d", 42)
		}, "boom %d", false},
		{"logger-errorf-injects", func(h *captureHook) {
			WithField("k", "v").Errorf(context.Background(), "boom")
		}, "boom", true},
		{"error-not-injects", func(h *captureHook) {
			Error(context.Background(), "plain args")
		}, "", false},
		{"infof-not-injects", func(h *captureHook) {
			Infof(context.Background(), "info %s", "x")
		}, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := installCaptureHook(t)
			c.log(h)
			if len(h.entries) != 1 {
				t.Fatalf("captured %d entries, want 1", len(h.entries))
			}
			data := h.entries[0].Data
			got, ok := data[ReservedProblemTypeField]
			if c.wantType == "" {
				if ok {
					t.Fatalf("reserved field should not be injected, got %v", got)
				}
				return
			}
			if !ok || got != c.wantType {
				t.Fatalf("reserved field = %v, want %q", got, c.wantType)
			}
			if c.wantExtra && data["k"] != "v" {
				t.Fatalf("WithField extra field lost: %v", data)
			}
		})
	}
}

// TestErrorfReservedFieldNotInLogLine 验证保留字段不泄漏进日志行：hook 未注册
// （无人消费即删）时由 custom formatter 输出侧兜底跳过。
func TestErrorfReservedFieldNotInLogLine(t *testing.T) {
	std := logrus.StandardLogger()
	origFormatter, origOut := std.Formatter, std.Out
	t.Cleanup(func() { std.Formatter, std.Out = origFormatter, origOut })

	var buf bytes.Buffer
	std.Formatter = NewCustomFormatter()
	std.Out = &buf

	Errorf(context.Background(), "boom %d", 42)

	out := buf.String()
	if !strings.Contains(out, "boom 42") {
		t.Fatalf("log line should contain formatted message, got %q", out)
	}
	if strings.Contains(out, ReservedProblemTypeField) {
		t.Fatalf("log line leaked reserved field: %q", out)
	}
}
