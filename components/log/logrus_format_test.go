package log

import (
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCustomFormatter_Format(t *testing.T) {
	var (
		formatter = NewCustomFormatter()
	)

	logrus.SetFormatter(formatter)
	logrus.SetReportCaller(true)

	defer func() {
		if r := recover(); r != nil {
			stack := make([]byte, 4096)
			stack = stack[:runtime.Stack(stack, false)]
			logrus.Errorf("[panic] err: %v, stack: %v", r, string(stack))
		}
	}()
	panic("test")
}
