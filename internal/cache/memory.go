package cache

import (
	"sync"
	"time"
)

type item struct {
	val       any
	expiresAt time.Time
}

// MemoryCache 是基于内存与 sync.RWMutex 的并发安全 TTL 缓存实现。
type MemoryCache struct {
	mu        sync.RWMutex
	items     map[string]item
	stopChan  chan struct{}
	ticker    *time.Ticker
	closeOnce sync.Once
}

// NewMemoryCache 创建一个新的内存缓存，并启动后台 goroutine 定期清理过期项。
func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	if cleanupInterval <= 0 {
		cleanupInterval = time.Minute
	}

	c := &MemoryCache{
		items:    make(map[string]item),
		stopChan: make(chan struct{}),
		ticker:   time.NewTicker(cleanupInterval),
	}

	go c.startCleanupLoop()

	return c
}

func (c *MemoryCache) startCleanupLoop() {
	for {
		select {
		case <-c.ticker.C:
			c.cleanupExpired()
		case <-c.stopChan:
			c.ticker.Stop()
			return
		}
	}
}

func (c *MemoryCache) cleanupExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, it := range c.items {
		if now.After(it.expiresAt) {
			delete(c.items, k)
		}
	}
}

// Get 获取缓存值。若 key 存在但已过期，会惰性删除并返回 (nil, false)。
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(it.expiresAt) {
		c.mu.Lock()
		// 二次确认，防止并发竞争
		if current, exists := c.items[key]; exists && time.Now().After(current.expiresAt) {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return it.val, true
}

// Set 存入缓存值并指定有效期 ttl。若 ttl <= 0，则默认保留 1 分钟。
func (c *MemoryCache) Set(key string, val any, ttl time.Duration) {
	if ttl <= 0 {
		ttl = time.Minute
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		val:       val,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete 删除指定 key。
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear 清空所有缓存项。
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]item)
}

// Close 关闭后台清理定时器并退出清理 goroutine。
func (c *MemoryCache) Close() {
	c.closeOnce.Do(func() {
		close(c.stopChan)
	})
}
