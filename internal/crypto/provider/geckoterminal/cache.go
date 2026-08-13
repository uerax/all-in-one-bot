package geckoterminal

import (
	"sync"
	"time"
)

type cacheItem struct {
	val       interface{}
	updatedAt time.Time
}

type ttlCache struct {
	mu          sync.RWMutex
	items       map[string]cacheItem
	ttlDuration time.Duration
}

func newTTLCache(ttl time.Duration) *ttlCache {
	return &ttlCache{
		items:       make(map[string]cacheItem),
		ttlDuration: ttl,
	}
}

func (c *ttlCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Since(item.updatedAt) > c.ttlDuration {
		return nil, false
	}
	return item.val, true
}

func (c *ttlCache) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{
		val:       val,
		updatedAt: time.Now(),
	}
}

// rateLimiter implements a simple token bucket for rate limiting requests (e.g., max 25 req/min).
type rateLimiter struct {
	mu           sync.Mutex
	capacity     int
	tokens       float64
	fillRate     float64 // tokens per second
	lastRefilled time.Time
}

func newRateLimiter(maxPerMin int) *rateLimiter {
	return &rateLimiter{
		capacity:     maxPerMin,
		tokens:       float64(maxPerMin),
		fillRate:     float64(maxPerMin) / 60.0,
		lastRefilled: time.Now(),
	}
}

func (r *rateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefilled).Seconds()
	r.lastRefilled = now

	r.tokens += elapsed * r.fillRate
	if r.tokens > float64(r.capacity) {
		r.tokens = float64(r.capacity)
	}

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		return true
	}

	return false
}
