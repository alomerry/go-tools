package cat

import (
	"context"
	"errors"
	"reflect"
	"strings"
)

// problemSeparator 为 problem message 内各段的连接符，与 ext 日志上报侧共用口径。
const problemSeparator = " · "

// fallbackProblemType 为 type tag 兜底值：错误类别推导失败时（err 全链为
// errors/fmt 的通用类型，或类型名解析失败）退化该固定值，保证点位仍可写入、
// 可按 type 过滤。
const fallbackProblemType = "error"

// LogError 记录一条异常打点（problem 点位）。args 作为附加信息（如 k=v 键值对）
// 并入 message，空段跳过。type tag 为从 err 推导的错误类别（对标 Java 异常类名，
// 如 PathError / boundsError），推导失败退化为 fallbackProblemType；可变细节
// 只进 message field，避免进 tag 造成 InfluxDB 序列基数膨胀。
func LogError(ctx context.Context, err error, args ...string) {
	LogErrorWithType(ctx, errorCategory(err), err, args...)
}

// LogErrorWithType 与 LogError 的差异仅在 type tag 来源：typ 由调用方显式指定
// 而非从 err 推导（logHook 场景——日志文本没有原始错误值，type 取 Errorf 注入
// 的调用点格式串或动态日志文本兜底）。typ 按 maxProblemType 截断（兜底形态
// 含可变长日志文本，须约束 tag 体积），截空或入参为空退化为 fallbackProblemType。
func LogErrorWithType(ctx context.Context, typ string, err error, args ...string) {
	if err == nil || !isEnabled() {
		return
	}

	// type tag 单行卫生：格式串与兜底日志文本都可能带换行，统一归一为空格
	// 后再截断。
	typ = truncateString(strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, typ), maxProblemType)
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

// errorCategory 沿 Unwrap 解包链取错误的具体类型名（reflect 取 Name，天然不含
// 包名与指针前缀，如 *fs.PathError → PathError）作为错误类别；panic 的运行时
// 错误（runtime.boundsError/divideError）由此自然产出语义类别。取链上首个具
// 语义类型名而非字面最内层：PathError/OpError 等包装类型必然继续 Unwrap 到
// Errno/哨兵，最内层口径会让类别整体丢失。errors/fmt 包内的通用包装产物按
// PkgPath 识别、仅作下钻节点（按裸类型名判定会误伤业务自定义同名类型，且追
// 不上 stdlib 新增），全链无语义类型时返回空串，由调用方兜底。限制：仅覆盖
// 单链 Unwrap() error，Unwrap() []error 形态的分支链（multi-%w 的 wrapErrors、
// errors.Join 产物）不支持下钻、整体落兜底 "error"，与标准库 errors.Is/As 的
// 单链遍历口径一致（当前两仓调用点无 Join 聚合，声明限制并锁定现状）。
func errorCategory(err error) string {
	for err != nil {
		t := reflect.TypeOf(err)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if pkg := t.PkgPath(); pkg != "errors" && pkg != "fmt" {
			if name := t.Name(); name != "" {
				return name
			}
		}
		err = errors.Unwrap(err)
	}
	return ""
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
