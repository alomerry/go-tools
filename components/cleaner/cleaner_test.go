package cleaner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCleaner_LIFOOrder(t *testing.T) {
	c := NewCleaner()

	var order []int
	var mu sync.Mutex

	// 注册 3 个清理函数
	for i := 1; i <= 3; i++ {
		val := i
		c.RegisterCleanFn(func(ctx context.Context) error {
			mu.Lock()
			order = append(order, val)
			mu.Unlock()
			return nil
		})
	}

	err := c.Clean(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 期望是 LIFO 执行顺序：3, 2, 1
	expected := []int{3, 2, 1}
	if len(order) != len(expected) {
		t.Fatalf("expected %d functions to run, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: expected %d, got %d", i, v, order[i])
		}
	}
}

func TestCleaner_Errors(t *testing.T) {
	c := NewCleaner()

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	c.RegisterCleanFn(func(ctx context.Context) error {
		return err1
	})
	c.RegisterCleanFn(func(ctx context.Context) error {
		return nil
	})
	c.RegisterCleanFn(func(ctx context.Context) error {
		return err2
	})

	err := c.Clean(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 期望收集到所有错误
	if !errors.Is(err, err1) {
		t.Errorf("expected error to contain err1")
	}
	if !errors.Is(err, err2) {
		t.Errorf("expected error to contain err2")
	}
}

func TestCleaner_ConcurrentRegistration(t *testing.T) {
	c := NewCleaner()

	var wg sync.WaitGroup
	const goroutines = 100

	// 并发注册
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.RegisterCleanFn(func(ctx context.Context) error {
				time.Sleep(1 * time.Millisecond) // 模拟轻量操作
				return nil
			})
		}()
	}

	wg.Wait()

	// 并发执行时的状态验证
	var count int32
	c.RegisterCleanFn(func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	err := c.Clean(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 虽然不知道具体顺序，但是应该执行了 101 个函数
	// 上面的 count 应该变成 1，因为只有一个修改了 count
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestCleaner_NilFunction(t *testing.T) {
	c := NewCleaner()

	// 注册 nil 函数不应 panic，并且不会添加到列表
	c.RegisterCleanFn(nil)

	err := c.Clean(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegisterCleanFn(t *testing.T) {
	var executed bool
	c := RegisterCleanFn(func(ctx context.Context) error {
		executed = true
		return nil
	})

	err := c.Clean(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !executed {
		t.Error("expected function to be executed")
	}
}
