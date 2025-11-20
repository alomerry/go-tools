package maps

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMap(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
		value, ok := c.Get("a")
		assert.True(t, ok)
		assert.Equal(t, 1, value)
	}()
	go func() {
		value, ok := c.Get("b")
		assert.True(t, ok)
		assert.Equal(t, 2, value)
	}()
	go func() {
		value, ok := c.Get("c")
		assert.True(t, ok)
		assert.Equal(t, 3, value)
	}()
	time.Sleep(time.Second)
	assert.Equal(t, 3, c.Size())
}

func TestConcurrentMap_Get(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
	}()
	go func() {
		c.Set("b", 2)
	}()
	go func() {
		c.Set("c", 3)
	}()
	time.Sleep(time.Second)
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

func TestConcurrentMap_Size(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
	}()
	go func() {
		c.Set("b", 2)
	}()
	go func() {
		c.Set("c", 3)
	}()
	time.Sleep(time.Second)
	assert.Equal(t, 3, c.Size())
}

func TestConcurrentMap_Clear(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
	}()
	go func() {
		c.Set("b", 2)
	}()
	go func() {
		c.Set("c", 3)
	}()
	time.Sleep(time.Second)
	c.Clear()
	assert.Equal(t, 0, c.Size())
}

func TestConcurrentMap_Keys(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
	}()
	go func() {
		c.Set("b", 2)
	}()
	go func() {
		c.Set("c", 3)
	}()
	time.Sleep(time.Second)
	assert.Equal(t, []string{"a", "b", "c"}, c.Keys())
}

func TestConcurrentMap_Values(t *testing.T) {
	c := ConcurrentMap[string, int]{}
	go func() {
		c.Set("a", 1)
	}()
	go func() {
		c.Set("b", 2)
	}()
	go func() {
		c.Set("c", 3)
	}()
	time.Sleep(time.Second)
	assert.Equal(t, []int{1, 2, 3}, c.Values())
}
