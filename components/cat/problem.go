package cat

import (
	"context"
	"fmt"
	"strings"
)

// LogError 记录一条异常打点（problem 点位）。args[0] 可指定分类，缺省用
// error 的类型名；其余 args 作为附加信息并入 message。type tag 固定为
// error，可变细节并入 message field，避免进 tag 造成 InfluxDB 序列基数膨胀。
func LogError(ctx context.Context, err error, args ...string) {
	if err == nil || !isEnabled() {
		return
	}

	var msg []string
	category := ""
	if len(args) > 0 {
		category = args[0]
		msg = args[1:]
	}
	if category == "" {
		category = fmt.Sprintf("%T", err)
	}

	parts := make([]string, 0, len(msg)+2)
	parts = append(parts, category)
	parts = append(parts, msg...)
	parts = append(parts, err.Error())

	emit(point{
		measurement: "problem",
		tags: map[string]string{
			"service": serviceName(),
			"type":    problemType,
		},
		fields: map[string]any{
			"message": truncateString(strings.Join(parts, " · "), maxProblemMessage),
		},
	})
}
