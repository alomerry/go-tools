package cat

import (
	"context"
	"strings"
)

// LogEvent 记录一条事件（event 点位：tags service/type/name，field data）。
// args 以换行拼接并入 data，总量截断到 maxEventData；未初始化时静默空操作。
func LogEvent(_ context.Context, mtype, name string, args ...string) {
	if !isEnabled() {
		return
	}

	emit(point{
		measurement: "event",
		tags: map[string]string{
			"service": serviceName(),
			"type":    mtype,
			"name":    name,
		},
		fields: map[string]any{
			"data": truncateString(strings.Join(args, "\n"), maxEventData),
		},
	})
}
