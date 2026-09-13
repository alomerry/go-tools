package cache

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLruCache_MissNotPanic(t *testing.T) {
	c := NewLru[string](context.Background(), WithCacheSize(4))
	// 修复前：miss 时 nil.(T) panic
	v, ok := c.Get("missing")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func TestLruCache_AddGet(t *testing.T) {
	c := NewLru[int](context.Background(), WithCacheSize(4))
	c.Add("a", 1)
	c.Add("b", 2)

	v, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)

	v, ok = c.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	// 容量淘汰：写入超过容量后最早的 key 应被淘汰且不 panic
	for i := 0; i < 10; i++ {
		c.Add("k"+string(rune('0'+i)), i)
	}
	_, ok = c.Get("a")
	assert.False(t, ok)
}

func TestLruCache_Concurrent(t *testing.T) {
	// 修复前：innerLru 无锁，并发 Add/Get 触发 fatal（race detector 可复现）
	c := NewLru[int](context.Background(), WithCacheSize(64))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				key := string(rune('a' + (base+j)%26))
				c.Add(key, base+j)
				c.Get(key)
			}
		}(i)
	}
	wg.Wait()
}
