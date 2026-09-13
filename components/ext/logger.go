package ext

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	_ "unsafe"

	"github.com/alomerry/go-tools/components/cat"
	"github.com/alomerry/go-tools/components/log"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func init() {
	Register(cons.ExtLogger, NewLoggerExt)
}

type loggerExt struct {
}

func NewLoggerExt() Ext {
	return loggerExt{}
}

func (loggerExt) Init(_ context.Context) error {
	logrus.SetFormatter(log.NewCustomFormatter())
	// caller 解析契约：logrus 内建 ReportCaller 刻意关闭（避免 custom formatter
	// 输出 caller 的格式抖动），entry.Caller 完全由 logHook.Fire 内 getCaller()
	// 自行注入，仅供 formatter 输出日志调用位置。依赖 entry.Caller 的路径必须
	// 容忍其为 nil（栈解析失败），严禁回退运行时 skip 解析——那会解析到
	// components/ext 包自身帧，污染日志调用位置的聚合。
	logrus.SetReportCaller(false)
	logrus.AddHook(logHook{})
	return nil
}

type logHook struct {
}

func (logHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (logHook) Fire(entry *logrus.Entry) error {
	if env.Local() {
		return nil
	}

	entry.Caller = getCaller()
	var extra []string
	for k, v := range entry.Data {
		extra = append(extra, fmt.Sprintf("%s=%v", k, cast.ToString(v)))
	}

	if entry.Level <= logrus.ErrorLevel {
		// type 统一取 entry.Message：logrus 已按调用点参数完成占位符格式化
		// （Errorf("not found: %s", path) → "not found: GET /ping"），带参与无参
		// 形态同一路径，告警侧展示真实日志文本而非未展开的 %s 格式串。注意
		// type 为动态文本（如 404 的任意 path 会产生新 tag），InfluxDB 序列基数
		// 由 cat 侧 maxProblemType 截断与调用方约束兜底，属已接受的取舍。
		reportProblem(entry.Context, entry.Message, errors.New(entry.Message), extra...)
		return nil
	}

	cat.AddData(entry.Context, strings.Join(append([]string{fmt.Sprintf("[%v]%s", entry.Level.String(), entry.Message)}, extra...), "\n"))
	return nil
}

// reportProblem 为 problem 打点出口。包级变量仅为单测注入捕获 type/err/extra，
// 生产行为固定指向 cat.LogErrorWithType。
var reportProblem = cat.LogErrorWithType

//go:linkname getPackageName github.com/sirupsen/logrus.getPackageName
func getPackageName(string) string

const (
	maximumCallerDepth int = 25
	minimumCallerDepth int = 4
)

func getCaller() *runtime.Frame {
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		if pkg == "github.com/sirupsen/logrus" || pkg == "github.com/alomerry/go-tools/components/log" {
			continue
		}

		return &f //nolint:scopelint
	}

	return nil
}
