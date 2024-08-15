package lexy

import "sync"

// A very simple, thread-safe, non-evicting cache which computes values.
// The compute function must be thread-safe, and should be idempotent.
// This cache only grows.
// This is hardly optimal, do not use for high-performance tasks.
// It can recompute the value for a key more than once if timing is unlucky.
type cache[K comparable, V any] struct {
	// pointer to prevent copying the mutex when this cache is passed by value
	lock    *sync.RWMutex
	cached  map[K]V
	compute func(K) V
}

func makeCache[K comparable, V any](compute func(K) V) cache[K, V] {
	checkNonNil(compute, "compute function")
	return cache[K, V]{
		lock:    &sync.RWMutex{},
		cached:  map[K]V{},
		compute: compute,
	}
}

func (c *cache[K, V]) getExisting(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.cached[key]
	return value, ok
}

func (c *cache[K, V]) Get(key K) V {
	value, ok := c.getExisting(key)
	if ok {
		return value
	}
	value = c.compute(key)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cached[key] = value
	return value
}
