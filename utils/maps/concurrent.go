package maps

import "sync"

type ConcurrentMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func (c *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.m[key]
	return value, ok
}

func (c *ConcurrentMap[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}

func (c *ConcurrentMap[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

func (c *ConcurrentMap[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

func (c *ConcurrentMap[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.m {
		delete(c.m, key)
	}
}

func (c *ConcurrentMap[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.m))
	for key := range c.m {
		keys = append(keys, key)
	}
	return keys
}

func (c *ConcurrentMap[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]V, 0, len(c.m))
	for _, value := range c.m {
		values = append(values, value)
	}
	return values
}