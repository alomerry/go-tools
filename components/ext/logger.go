package ext

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	_ "unsafe"

	"github.com/alomerry/cat-go/cat"
	"github.com/alomerry/go-tools/components/log"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// logSeparate is the separator prepended to error messages reported to cat.
// Callers may override it via SetLogSeparate before LoadExt.
var logSeparate = "·"

// SetLogSeparate overrides the separator used when reporting errors to cat.
func SetLogSeparate(s string) {
	if s != "" {
		logSeparate = s
	}
}

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
		cat.LogErrorWithCategoryBySkipTrace(entry.Context, errors.New(logSeparate+entry.Message), entry.Message, 9, extra...)
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
