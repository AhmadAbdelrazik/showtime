package cache

import (
	"context"
	"sync"
	"time"
)

type Cache struct {
	data          map[string]*cacheData
	mu            sync.RWMutex
	cancelCleanup context.CancelFunc
}

type cacheData struct {
	value     string
	lastCheck time.Time
}

func New() *Cache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &Cache{
		data:          make(map[string]*cacheData),
		cancelCleanup: cancel,
	}

	go cache.clean(ctx)

	return cache
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, ok := c.data[key]
	if !ok {
		return ""
	}

	c.mu.Lock()
	data.lastCheck = time.Now()
	c.mu.Unlock()

	return data.value
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheData{
		value:     value,
		lastCheck: time.Now(),
	}
}

func (c *Cache) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

func (c *Cache) clean(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			for k, v := range c.data {
				if time.Since(v.lastCheck) > 15*time.Minute {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Stop the cache cleanup function
func (c *Cache) Stop() {
	c.cancelCleanup()
}
