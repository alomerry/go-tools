package ext

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/log"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/sirupsen/logrus"
)

// problemCall 捕获 reportProblem（cat.LogErrorWithType 出口）的 type 入参。
type problemCall struct {
	typ string
}

// installProblemRecorder 以 reportProblem 包级变量为 seam 捕获 problem 打点，
// 测试结束恢复。cat 未初始化时 isEnabled 恒 false，经真实 cat 链路无法观测，
// 故 seam 注入是唯一直观测点。
func installProblemRecorder(t *testing.T) *[]problemCall {
	t.Helper()
	calls := &[]problemCall{}
	orig := reportProblem
	reportProblem = func(_ context.Context, typ string, _ error, _ ...string) {
		*calls = append(*calls, problemCall{typ: typ})
	}
	t.Cleanup(func() { reportProblem = orig })
	return calls
}

// installLogHook 注册 logHook 并保存 logrus 全局状态，测试结束恢复，避免用例
// 间串扰。
func installLogHook(t *testing.T) {
	t.Helper()
	std := logrus.StandardLogger()
	origHooks := make(logrus.LevelHooks, len(std.Hooks))
	for lvl, hooks := range std.Hooks {
		origHooks[lvl] = append([]logrus.Hook(nil), hooks...)
	}
	origOut := std.Out
	t.Cleanup(func() {
		std.Hooks, std.Out = origHooks, origOut
	})
	logrus.AddHook(logHook{})
	logrus.SetOutput(&nopWriter{})
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

// TestLogHookProblemTypeIsFormattedMessage 验证 problem type 新语义：统一取
// logrus 已按调用点参数格式化的 entry.Message（真实 logrus 调用驱动），带参与
// 无参形态同一路径——修复「not found: %s %s」未格式化占位符直透 kook 告警的
// 问题；空文本原样透传，由 cat 侧 fallbackProblemType 兜底（cat 包测试覆盖）。
// Fatal/Panic 级与 Error 共用 hook 的 entry.Level <= ErrorLevel 分支，不重复
// 驱动（Fatalf 会 os.Exit、Panicf 会 panic，手工构造 entry 无额外信息量）。
func TestLogHookProblemTypeIsFormattedMessage(t *testing.T) {
	t.Setenv(cons.ENV, "prod")
	calls := installProblemRecorder(t)
	installLogHook(t)

	log.Errorf(context.Background(), "not found: %s %s", "GET", "/ping1")
	log.Error(context.Background(), "plain text")
	log.Error(context.Background(), "")

	want := []problemCall{
		{typ: "not found: GET /ping1"},
		{typ: "plain text"},
		{typ: ""},
	}
	if len(*calls) != len(want) {
		t.Fatalf("captured %d problem calls, want %d: %+v", len(*calls), len(want), *calls)
	}
	for i, c := range *calls {
		if c.typ != want[i].typ {
			t.Fatalf("call[%d] type = %q, want %q", i, c.typ, want[i].typ)
		}
	}
}

// TestLogHookNonErrorSkipsProblem 验证 Info/Warn 级别不打 problem 点位（不打
// 点位即无告警噪音），仅走 AddData 聚合路径。
func TestLogHookNonErrorSkipsProblem(t *testing.T) {
	t.Setenv(cons.ENV, "prod")
	calls := installProblemRecorder(t)
	installLogHook(t)

	log.Info(context.Background(), "info text")

	if len(*calls) != 0 {
		t.Fatalf("non-error level should not report problem, got %+v", *calls)
	}
}
