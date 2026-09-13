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
	// Errorf 注入的 problem type 保留字段只进 problem tag：先取后删，保证它
	// 不进下方 extras 遍历、也不残留给 formatter（含 env.Local 短路路径——
	// 本地日志行同样不能出现该内部字段）。
	typ, _ := entry.Data[log.ReservedProblemTypeField].(string)
	delete(entry.Data, log.ReservedProblemTypeField)

	if env.Local() {
		return nil
	}

	entry.Caller = getCaller()
	var extra []string
	for k, v := range entry.Data {
		extra = append(extra, fmt.Sprintf("%s=%v", k, cast.ToString(v)))
	}

	if entry.Level <= logrus.ErrorLevel {
		// type 优先取 Errorf 注入的调用点格式串（编译期常量，可聚合、可反查
		// 源码）；无格式串形态（log.Error(args...)）退化为动态日志文本，cat
		// 侧截断后作 tag，文本也为空才落 fallbackProblemType。message 只保留
		// 一份日志原文（原文 · k1=v1 · k2=v2）。
		if typ == "" {
			typ = entry.Message
		}
		cat.LogErrorWithType(entry.Context, typ, errors.New(entry.Message), extra...)
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
