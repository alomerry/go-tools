package log

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils"
	time2 "github.com/alomerry/go-tools/utils/time"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type customFormatter struct {
}

var (
	d8, _ = time.LoadLocation("Asia/Shanghai")
)

func (c *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    var (
        buffer = *entry.Buffer

        module  = "-"
        traceId = "-"
    )

	if entry.Caller != nil {
		module = fmt.Sprintf("%s:%s:%d", path.Base(entry.Caller.File), entry.Caller.Function, entry.Caller.Line)
	}

	if entry.Context != nil {
		traceId = utils.FromCtx(entry.Context, cons.TraceIdKey)
	} else {
		traceId = uuid.New().String()
	}

	fmt.Fprintf(&buffer, "[%s]•[%s]•[%s]:[%s]",
		entry.Time.Format(time2.Readable),
		strings.ToUpper(entry.Level.String()),
		module,
		traceId,
	)

	for key, value := range entry.Data {
		fmt.Fprintf(&buffer, "[%s:%v]", key, value)
	}

	// 保证日志条目只占一行：将消息中的换行、回车替换为转义字符
	msg := strings.ReplaceAll(entry.Message, "\r\n", "\\n")
	msg = strings.ReplaceAll(msg, "\n", "\\n")
	msg = strings.ReplaceAll(msg, "\r", "\\r")
	fmt.Fprintf(&buffer, "%s", msg)
	// 统一追加单个换行作为日志条目的结尾
	buffer.WriteString("\n")

    return buffer.Bytes(), nil
}

func NewCustomFormatter() logrus.Formatter {
	return &customFormatter{}
}
