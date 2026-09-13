package maps

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMap(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for i, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string, val int) {
			defer wg.Done()
			c.Set(key, val)
			value, ok := c.Get(key)
			assert.True(t, ok)
			assert.Equal(t, val, value)
		}(k, i+1)
	}
	wg.Wait()
	assert.Equal(t, 3, c.Size())
}

func TestConcurrentMap_Get(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for i, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string, val int) {
			defer wg.Done()
			c.Set(key, val)
		}(k, i+1)
	}
	wg.Wait()
	assert.Equal(t, 3, c.Size())
	value, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
	value, ok = c.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, value)
	value, ok = c.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, value)
}

func TestConcurrentMap_ZeroValue(t *testing.T) {
	// 零值直接可用：Set 懒初始化，Get/Delete/Size 对 nil map 安全
	var c ConcurrentMap[string, int]
	_, ok := c.Get("missing")
	assert.False(t, ok)
	assert.Equal(t, 0, c.Size())
	c.Set("a", 1)
	value, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
	c.Delete("a")
	assert.Equal(t, 0, c.Size())
}

func TestConcurrentMap_Size(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for i, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string, val int) {
			defer wg.Done()
			c.Set(key, val)
		}(k, i+1)
	}
	wg.Wait()
	assert.Equal(t, 3, c.Size())
}

func TestConcurrentMap_Clear(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for i, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string, val int) {
			defer wg.Done()
			c.Set(key, val)
		}(k, i+1)
	}
	wg.Wait()
	c.Clear()
	assert.Equal(t, 0, c.Size())
}

func TestConcurrentMap_Keys(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			c.Set(key, 1)
		}(k)
	}
	wg.Wait()
	// map 无序，按集合比较
	assert.ElementsMatch(t, []string{"a", "b", "c"}, c.Keys())
}

func TestConcurrentMap_Values(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	var wg sync.WaitGroup
	for i, k := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string, val int) {
			defer wg.Done()
			c.Set(key, val)
		}(k, i+1)
	}
	wg.Wait()
	assert.ElementsMatch(t, []int{1, 2, 3}, c.Values())
}
