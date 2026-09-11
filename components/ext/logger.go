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
	// 自行注入。依赖 entry.Caller 的路径（logHook problem 打点等）必须容忍其为
	// nil（栈解析失败）并走兜底口径，严禁回退运行时 skip 解析——那会解析到
	// components/ext 包自身帧，污染 problem type 聚合。
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
		// type 统一为真实业务调用方的完整包路径（与直接 cat.LogError 口径一致）。
		// hook 栈帧不代表业务调用方，故显式透传 entry.Caller 解析出的包路径；
		// entry.Caller 为 nil（getCaller 栈解析失败）时直接用 FallbackProblemType
		// 兜底——若回退运行时 skip 解析会命中 ext 包自身帧，污染聚合口径。
		// message 只保留一份日志原文（原文 · k1=v1 · k2=v2），hook 不再重复传原文。
		typ := cat.FallbackProblemType
		if entry.Caller != nil {
			typ = cat.PackagePath(entry.Caller.Function)
		}
		cat.LogErrorWithCaller(entry.Context, typ, errors.New(entry.Message), extra...)
		return nil
	}

	cat.AddData(entry.Context, strings.Join(append([]string{fmt.Sprintf("[%v]%s", entry.Level.String(), entry.Message)}, extra...), "\n"))
	return nil
}

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
