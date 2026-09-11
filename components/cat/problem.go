package cat

import (
	"context"
	"runtime"
	"strings"
)

// problemSeparator 为 problem message 内各段的连接符，与 ext 日志上报侧共用口径。
const problemSeparator = " · "

// FallbackProblemType 为 type tag 兜底值：调用方包路径解析失败时（极端场景，
// 如 runtime 栈不可用）退化为旧的固定值，保证点位仍可写入、可按 type 过滤。
// 导出供 ext 日志上报侧共用同一兜底口径（见 components/ext/logger.go logHook）。
const FallbackProblemType = "error"

// fallbackProblemType 包内别名，保持既有引用简短。
const fallbackProblemType = FallbackProblemType

// LogError 记录一条异常打点（problem 点位）。args 作为附加信息（如 k=v 键值对）
// 并入 message，空段跳过。type tag 为调用方的完整包路径
// （如 github.com/alomerry/homelab-backend/service/blog/model），可直接 cat 调用
// 与 logrus hook 两条路径统一口径；可变细节只进 message field，避免进 tag 造成
// InfluxDB 序列基数膨胀。
func LogError(ctx context.Context, err error, args ...string) {
	// 在 LogError 自身帧上直接取 runtime.Caller(1)：1=LogError 的调用方，即真实
	// 业务调用方。不经 callerPackage(skip) 间接链（"LogError → LogErrorWithCaller
	// → callerPackage(2)" 的跨函数 skip 推算随包装层级漂移，历史上曾错指到
	// LogError/包装函数自身帧，type 退化为 components/cat 包路径污染聚合口径），
	// 解析失败传空串由 LogErrorWithCaller 统一兜底。
	typ := ""
	if pc, _, _, ok := runtime.Caller(1); ok {
		typ = PackagePath(runtime.FuncForPC(pc).Name())
	}
	LogErrorWithCaller(ctx, typ, err, args...)
}

// LogErrorWithCaller 与 LogError 相同，但 type tag 由调用方显式指定（典型为
// logrus hook：hook 栈帧不代表真实业务调用方，须透传其预先解析的调用方包路径）。
// typ 为空时按运行时解析直接调用方的包路径（skip：0=callerPackage、
// 1=LogErrorWithCaller、2=直接调用方），再失败则退化为 FallbackProblemType。
func LogErrorWithCaller(ctx context.Context, typ string, err error, args ...string) {
	if err == nil || !isEnabled() {
		return
	}

	if typ == "" {
		typ = callerPackage(2)
	}
	if typ == "" {
		typ = fallbackProblemType
	}

	emit(point{
		measurement: "problem",
		tags: map[string]string{
			"service": serviceName(),
			"type":    typ,
		},
		fields: map[string]any{
			"message": truncateString(joinMessage(err.Error(), args...), maxProblemMessage),
		},
	})
}

// callerPackage 返回运行时调用方（skip 口径同 runtime.Caller）的完整包路径。
// 解析失败返回空串，由调用方决定兜底。
func callerPackage(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return PackagePath(runtime.FuncForPC(pc).Name())
}

// PackagePath 从 runtime 函数全名（如 github.com/foo/bar/pkg.(*T).M）解析出
// 完整包路径（github.com/foo/bar/pkg）。无法识别时原样返回。
func PackagePath(fn string) string {
	slash := strings.LastIndex(fn, "/")
	dot := strings.Index(fn[slash+1:], ".")
	if dot < 0 {
		return fn
	}
	return fn[:slash+1+dot]
}

// joinMessage 以 problemSeparator 连接各段，跳过空段。曾在此做跨段去重
// （logrus hook 侧 category 与 err 传同一份日志原文），现 hook 已改为只传
// 一份原文 + k=v extras，去重属对已修复 bug 的防御性复杂度，随之移除。
func joinMessage(first string, rest ...string) string {
	parts := make([]string, 0, len(rest)+1)
	for _, p := range append([]string{first}, rest...) {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, problemSeparator)
}
