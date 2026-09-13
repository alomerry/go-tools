package log

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/alomerry/go-tools/components/cat"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils"
	time2 "github.com/alomerry/go-tools/utils/time"
	"github.com/alomerry/go-tools/utils/trace"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type customFormatter struct {
}

var (
	d8, _ = time.LoadLocation("Asia/Shanghai")
)

func (c *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var (
		// 必须用指针：按值拷贝 bytes.Buffer 后写入拷贝会绕开 pool 状态
		buffer = entry.Buffer

		module  = "-"
		traceId = trace.GetTraceId(entry.Context, "-")
	)

	if entry.Caller != nil {
		module = fmt.Sprintf("%s:%s:%d", path.Base(entry.Caller.File), entry.Caller.Function, entry.Caller.Line)
	}

	if entry.Context != nil {
		traceId = utils.FromCtx(entry.Context, cons.TraceIdKey)
		if len(traceId) == 0 {
			traceId = cat.TraceIdFromCtx(entry.Context)
		}
	}

	if len(traceId) == 0 {
		traceId = xid.New().String()
	}

	_, _ = fmt.Fprintf(buffer, "[%s]•[%s]•[%s]:[%s]",
		// 统一上海时区输出（容器内多为 UTC，d8 曾加载后未使用导致日志差 8h）
		entry.Time.In(d8).Format(time2.Readable),
		strings.ToUpper(entry.Level.String()),
		module,
		traceId,
	)

	for key, value := range entry.Data {
		_, _ = fmt.Fprintf(buffer, "[%s:%v]", key, value)
	}

	// 保证日志条目只占一行：将消息中的换行、回车替换为转义字符
	msg := strings.ReplaceAll(entry.Message, "\r\n", "\\n")
	msg = strings.ReplaceAll(msg, "\n", "\\n")
	msg = strings.ReplaceAll(msg, "\r", "\\r")
	_, _ = fmt.Fprintf(buffer, "%s", msg)
	// 统一追加单个换行作为日志条目的结尾
	buffer.WriteString("\n")

	return buffer.Bytes(), nil
}

func NewCustomFormatter() logrus.Formatter {
	return &customFormatter{}
}
