package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

// ReservedProblemTypeField 为 Errorf/Fatalf/Panicf 注入 problem type 的保留字段
// 名：格式串经 WithField 挂入 entry.Data，ext logHook 取之作 problem 点位的
// type tag（编译期常量，天然有界、可聚合、可反查源码）。字段只在 log 与 ext 间
// 传递、不属日志内容——hook 消费后即删、custom formatter 输出时跳过，保证不
// 泄漏进日志行；业务调用方勿以此名作为 WithField key，字段会被 hook 吞掉。
const ReservedProblemTypeField = "__problemType"

func WithField(key string, value interface{}) Logger {
	return Logger{logrus.WithField(key, value)}
}

func WithFields(fields logrus.Fields) Logger {
	return Logger{logrus.WithFields(fields)}
}

func (l Logger) Info(ctx context.Context, args ...any) {
	l.WithContext(ctx).Info(args...)
}

func (l Logger) Warn(ctx context.Context, args ...any) {
	l.WithContext(ctx).Warn(args...)
}

func (l Logger) Error(ctx context.Context, args ...any) {
	l.WithContext(ctx).Error(args...)
}

func (l Logger) Panic(ctx context.Context, args ...any) {
	l.WithContext(ctx).Panic(args...)
}

func (l Logger) Fatal(ctx context.Context, args ...any) {
	l.WithContext(ctx).Fatal(args...)
}

func (l Logger) Infof(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).Infof(format, args...)
}

func (l Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).Warnf(format, args...)
}

// Errorf 注入格式串供 logHook 产出 problem type。级别序 Panic(0) < Fatal(1) <
// Error(2)，ext hook 的 entry.Level <= ErrorLevel 对三者统一走 error 级 problem
// 路径，故 Fatalf/Panicf 同口径注入；Info/Warn 不打 problem 点位，Error 的无格
// 式串形态没有调用点常量可注入，由 hook 侧按动态文本兜底。
func (l Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).WithField(ReservedProblemTypeField, format).Errorf(format, args...)
}

func (l Logger) Panicf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).WithField(ReservedProblemTypeField, format).Panicf(format, args...)
}

func (l Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).WithField(ReservedProblemTypeField, format).Fatalf(format, args...)
}

func Info(ctx context.Context, args ...any) {
	logrus.WithContext(ctx).Info(args...)
}

func Warn(ctx context.Context, args ...any) {
	logrus.WithContext(ctx).Warn(args...)
}

func Error(ctx context.Context, args ...any) {
	logrus.WithContext(ctx).Error(args...)
}

func Panic(ctx context.Context, args ...any) {
	logrus.WithContext(ctx).Panic(args...)
}

func Fatal(ctx context.Context, args ...any) {
	logrus.WithContext(ctx).Fatal(args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).Infof(format, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	// 语义同 Logger.Errorf：格式串作保留字段注入。
	logrus.WithContext(ctx).WithField(ReservedProblemTypeField, format).Errorf(format, args...)
}

func Panicf(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).WithField(ReservedProblemTypeField, format).Panicf(format, args...)
}

func Fatalf(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).WithField(ReservedProblemTypeField, format).Fatalf(format, args...)
}
