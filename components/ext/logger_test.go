package ext

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/components/log"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/sirupsen/logrus"
)

// TestLogHookStripsReservedProblemTypeField 验证 hook 取走保留字段后即从
// entry.Data 删除：字段不进 extras、不残留给 formatter 输出；用户附加字段
// 不被误删。local case 钉住「删除发生在 env.Local 短路之前」（本地日志行同样
// 不能出现该内部字段）；non-local case 防护「先取后删」先于 extras 组装的重构
// 漂移（cat 未初始化时点位行为由 cat 包测试覆盖）。
func TestLogHookStripsReservedProblemTypeField(t *testing.T) {
	cases := []struct{ name, env string }{
		{"local-short-circuit", cons.EnvLocal},
		{"non-local", "prod"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(cons.ENV, c.env)

			entry := logrus.WithField(log.ReservedProblemTypeField, "boom %d").WithField("k", "v")
			entry.Level = logrus.ErrorLevel
			entry.Message = "boom 42"
			entry.Context = context.Background()

			if err := (logHook{}).Fire(entry); err != nil {
				t.Fatalf("Fire returned error: %v", err)
			}
			if _, ok := entry.Data[log.ReservedProblemTypeField]; ok {
				t.Fatal("reserved field should be consumed and removed from entry.Data")
			}
			if entry.Data["k"] != "v" {
				t.Fatalf("user extra field lost: %v", entry.Data)
			}
		})
	}
}
