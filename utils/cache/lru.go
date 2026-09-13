package cache

import (
	"context"
	"sync"
	"time"

	"container/heap"

	"github.com/golang/groupcache/lru"
)

type Lru[T any] interface {
	Add(key string, value T)
	Get(key string) (T, bool)
}

type lruCache[T any] struct {
	innerLru *lru.Cache
	heap     heap.Interface

	// expires 记录各 key 的过期时间点（maxTTL>0 时启用 TTL 语义）
	expires map[string]time.Time
	maxTTL  time.Duration

	// 内嵌锁：groupcache lru 的 Get 命中会更新内部访问堆（非纯读），
	// Add/Get 统一使用写锁，故仅暴露互斥语义
	sync.Mutex
}

type LruOption func(*lruCacheOptions)

type lruCacheOptions struct {
	size   int
	maxTTL time.Duration
}

func WithCacheSize(size int) LruOption {
	return func(l *lruCacheOptions) {
		l.size = size
	}
}

func WithCacheMaxTTL(maxTTL string) LruOption {
	return func(l *lruCacheOptions) {
		ttl, err := time.ParseDuration(maxTTL)
		if err != nil {
			// 解析失败保持默认值（10min），不静默归零
			return
		}
		l.maxTTL = ttl
	}
}

func NewLru[T any](ctx context.Context, opts ...LruOption) Lru[T] {
	cache := &lruCache[T]{
		expires: make(map[string]time.Time),
	}

	option := &lruCacheOptions{
		size:   1000,
		maxTTL: 10 * time.Minute,
	}

	for _, opt := range opts {
		opt(option)
	}

	cache.innerLru = lru.New(option.size)
	cache.maxTTL = option.maxTTL
	cache.init(ctx)
	return cache
}

// init 启动过期清理 goroutine：随 ctx 取消退出（原实现永不退出且无条件
// RemoveOldest）。maxTTL<=0 时 LRU 容量淘汰已足够，不启动。
func (l *lruCache[T]) init(ctx context.Context) {
	if l.maxTTL <= 0 {
		return
	}
	interval := l.maxTTL / 2
	if interval > 10*time.Second {
		interval = 10 * time.Second
	}
	if interval < time.Second {
		interval = time.Second
	}
	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				l.removeExpired()
			}
		}
	}()
}

func (l *lruCache[T]) removeExpired() {
	now := time.Now()
	l.Lock()
	defer l.Unlock()
	for key, exp := range l.expires {
		if now.After(exp) {
			l.innerLru.Remove(key)
			delete(l.expires, key)
		}
	}
}

func (l *lruCache[T]) Add(key string, value T) {
	l.Lock()
	defer l.Unlock()
	l.innerLru.Add(key, value)
	if l.maxTTL > 0 {
		l.expires[key] = time.Now().Add(l.maxTTL)
	}
}

// Get 使用写锁：groupcache lru 命中时会更新内部访问堆（提升新鲜度），并非纯读。
func (l *lruCache[T]) Get(key string) (T, bool) {
	l.Lock()
	defer l.Unlock()
	if l.maxTTL > 0 {
		if exp, ok := l.expires[key]; ok && time.Now().After(exp) {
			l.innerLru.Remove(key)
			delete(l.expires, key)
		}
	}
	val, exists := l.innerLru.Get(key)
	if !exists {
		var zero T
		return zero, false
	}
	return val.(T), true
}
