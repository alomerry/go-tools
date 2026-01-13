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

	sync.RWMutex
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
		ttl, _ := time.ParseDuration(maxTTL)
		l.maxTTL = ttl
	}
}

func NewLru[T any](ctx context.Context, opts ...LruOption) Lru[T] {
	cache := &lruCache[T]{}

	option := &lruCacheOptions{
		size:   1000,
		maxTTL: 10 * time.Minute,
	}

	for _, opt := range opts {
		opt(option)
	}

	cache.innerLru = lru.New(option.size)
	cache.init()
	return cache
}

func (l *lruCache[T]) init() {
	go func() {
		tick := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-tick.C:
				l.innerLru.Get()
			}
		}
	}()
}

func (l *lruCache[T]) Add(key string, value T) {
	l.innerLru.Add(key, value)
}

func (l *lruCache[T]) Get(key string) (T, bool) {
	val, exists := l.innerLru.Get(key)
	return val.(T), exists
}
