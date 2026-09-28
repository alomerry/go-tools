package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

func WithField(key string, value interface{}) Logger {
	return Logger{logrus.WithField(key, value)}
}

func WithFields(fields logrus.Fields) Logger {
	return Logger{logrus.WithFields(fields)}
}

// WithError 把 err 挂到 entry.Data["error"]，不进 Errorf 的 format/args；nil 时不注入字段。
func WithError(err error) Logger {
	if err == nil {
		return Logger{logrus.WithFields(logrus.Fields{})}
	}
	return Logger{logrus.WithError(err)}
}

// Logger 须由包级 With* 构造（先 With 后 Errorf，Errorf 返回 void 无法再链）。
func (l Logger) WithField(key string, value interface{}) Logger {
	return Logger{l.Entry.WithField(key, value)}
}

func (l Logger) WithFields(fields logrus.Fields) Logger {
	return Logger{l.Entry.WithFields(fields)}
}

func (l Logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return Logger{l.Entry.WithError(err)}
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

// 低基数规范：format 只传稳定常量，可变细节走 WithField/WithError
// （如 log.WithError(err).WithField("path", p).Errorf(ctx, "not found")）；老调用（format 带变量）仍按 logrus 语义展开。
func (l Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).Errorf(format, args...)
}

func (l Logger) Panicf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).Panicf(format, args...)
}

func (l Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.WithContext(ctx).Fatalf(format, args...)
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
	logrus.WithContext(ctx).Errorf(format, args...)
}

func Panicf(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).Panicf(format, args...)
}

func Fatalf(ctx context.Context, format string, args ...any) {
	logrus.WithContext(ctx).Fatalf(format, args...)
}
