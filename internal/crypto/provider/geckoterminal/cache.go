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

// rateLimiter implements a token bucket with pacing control for rate limiting requests.
type rateLimiter struct {
	mu           sync.Mutex
	capacity     int
	tokens       float64
	fillRate     float64 // tokens per second
	lastRefilled time.Time
	lastRequest  time.Time
	minInterval  time.Duration // 两次请求间的最小安全间隔，防突发
}

func newRateLimiter(maxPerMin int, minInterval time.Duration) *rateLimiter {
	capVal := 3
	if maxPerMin < capVal {
		capVal = maxPerMin
	}
	return &rateLimiter{
		capacity:     capVal,
		tokens:       float64(capVal),
		fillRate:     float64(maxPerMin) / 60.0,
		lastRefilled: time.Now(),
		lastRequest:  time.Now().Add(-minInterval),
		minInterval:  minInterval,
	}
}

// Wait 阻塞直至获取可用令牌，并确保距离上次请求已满足 minInterval 间隔，杜绝瞬间突发流量。
func (r *rateLimiter) Wait() {
	r.mu.Lock()
	now := time.Now()

	// 1. 防突发最小间隔检查：距离上次请求至少过去 minInterval
	timeSinceLast := now.Sub(r.lastRequest)
	var sleepPacing time.Duration
	if timeSinceLast < r.minInterval {
		sleepPacing = r.minInterval - timeSinceLast
	}

	// 2. 令牌桶检查
	elapsed := now.Sub(r.lastRefilled).Seconds()
	r.lastRefilled = now
	r.tokens += elapsed * r.fillRate
	if r.tokens > float64(r.capacity) {
		r.tokens = float64(r.capacity)
	}

	var sleepToken time.Duration
	if r.tokens < 1.0 {
		missing := 1.0 - r.tokens
		sleepToken = time.Duration(missing/r.fillRate*float64(time.Second)) + 10*time.Millisecond
	}

	sleepTime := sleepPacing
	if sleepToken > sleepTime {
		sleepTime = sleepToken
	}

	if sleepTime > 0 {
		r.mu.Unlock()
		time.Sleep(sleepTime)
		r.mu.Lock()
		// 醒来后重新核算令牌
		now = time.Now()
		elapsed = now.Sub(r.lastRefilled).Seconds()
		r.lastRefilled = now
		r.tokens += elapsed * r.fillRate
		if r.tokens > float64(r.capacity) {
			r.tokens = float64(r.capacity)
		}
	}

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
	} else {
		r.tokens = 0
	}
	r.lastRequest = time.Now()
	r.mu.Unlock()
}

func (r *rateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastRequest) < r.minInterval {
		return false
	}

	elapsed := now.Sub(r.lastRefilled).Seconds()
	r.lastRefilled = now

	r.tokens += elapsed * r.fillRate
	if r.tokens > float64(r.capacity) {
		r.tokens = float64(r.capacity)
	}

	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		r.lastRequest = now
		return true
	}

	return false
}
