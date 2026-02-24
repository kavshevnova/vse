package main

// Задача: простой in-memory cache.

import (
	"fmt"
	"sync"
)

type Cache interface {
	Set(k, v string)
	Get(k string) (v string, ok bool)
}

// inMemoryCache — потокобезопасная реализация Cache.
type inMemoryCache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() Cache {
	return &inMemoryCache{data: make(map[string]string)}
}

func (c *inMemoryCache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = v
}

func (c *inMemoryCache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[k]
	return v, ok
}

func main() {
	c := NewCache()
	c.Set("name", "Alice")
	v, ok := c.Get("name")
	fmt.Println(v, ok) // Alice true
	_, ok2 := c.Get("missing")
	fmt.Println(ok2) // false
}
