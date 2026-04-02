package cleaner

import (
	"context"
	"errors"
	"sync"
)

type CleanFn func(context.Context) error

type Cleaner interface {
	RegisterCleanFn(CleanFn) Cleaner
	Clean(context.Context) error
}

var _ Cleaner = (*cleaner)(nil)

type cleaner struct {
	mu       sync.Mutex
	cleanFns []CleanFn
}

// RegisterCleanFn 注册清理函数。
// 支持并发调用，且会忽略 nil 的函数。
func (c *cleaner) RegisterCleanFn(fn CleanFn) Cleaner {
	if fn == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanFns = append(c.cleanFns, fn)
	return c
}

// Clean 执行所有已注册的清理操作。
// 按照 LIFO (后进先出) 顺序执行，与 defer 行为保持一致。
func (c *cleaner) Clean(ctx context.Context) error {
	c.mu.Lock()
	fns := make([]CleanFn, len(c.cleanFns))
	copy(fns, c.cleanFns)
	c.mu.Unlock()

	var errs []error
	// LIFO 倒序执行
	for i := len(fns) - 1; i >= 0; i-- {
		if err := fns[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	// errors.Join 自动处理无错误时返回 nil 的逻辑
	return errors.Join(errs...)
}

// NewCleaner 创建并返回一个新的 Cleaner 实例
func NewCleaner() Cleaner {
	return &cleaner{}
}

// RegisterCleanFn 作为构建器入口，创建一个新的 cleaner 并注册第一个函数
func RegisterCleanFn(fn CleanFn) Cleaner {
	return NewCleaner().RegisterCleanFn(fn)
}
