package main

import (
	"fmt"
	"sync"
	"time"
)

// ТЗ: Система кэширования с поддержкой TTL и конкурентного доступа.
// Множество горутин должны безопасно читать и писать в кэш.

type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

type Cache struct {
	mu    sync.Mutex
	items map[string]CacheItem
	stats CacheStats
}

type CacheStats struct {
	Hits   int
	Misses int
	mu     sync.Mutex
}

func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
		stats: CacheStats{},
	}
	go cache.cleanupExpired()
	return cache
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	return item.Value, true
}

func (c *Cache) GetStats() CacheStats {
	return c.stats
}

func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) GetAll() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]interface{})
	now := time.Now()

	for key, item := range c.items {
		if now.Before(item.ExpiresAt) {
			result[key] = item.Value
		}
	}

	return result
}

func (c *Cache) UpdateStats(key string, found bool) {
	if found {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}

func (c *Cache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func main() {
	cache := NewCache()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Set(key, fmt.Sprintf("value_%d_%d", id, j), 5*time.Second)
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				key := fmt.Sprintf("key_%d_%d", id%100, j%10)
				value, found := cache.Get(key)
				if found {
					fmt.Printf("Reader %d found: %s\n", id, value)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				stats := cache.GetStats()
				fmt.Printf("Stats %d: Hits=%d, Misses=%d\n", id, stats.Hits, stats.Misses)
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				all := cache.GetAll()
				fmt.Printf("GetAll %d: size=%d\n", id, len(all))
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("Final cache size: %d\n", cache.Size())
	finalStats := cache.GetStats()
	fmt.Printf("Final stats: Hits=%d, Misses=%d\n", finalStats.Hits, finalStats.Misses)
}
