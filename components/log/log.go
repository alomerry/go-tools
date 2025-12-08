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
