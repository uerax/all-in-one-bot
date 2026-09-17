package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryCache_BasicCRUD(t *testing.T) {
	c := NewMemoryCache(100 * time.Millisecond)
	defer c.Close()

	// 1. Get nonexistent
	if _, ok := c.Get("nonexistent"); ok {
		t.Error("expected false for nonexistent key")
	}

	// 2. Set & Get
	c.Set("foo", "bar", 1*time.Second)
	val, ok := c.Get("foo")
	if !ok || val != "bar" {
		t.Fatalf("expected 'bar', got %v (ok: %t)", val, ok)
	}

	// 3. Delete
	c.Delete("foo")
	if _, ok := c.Get("foo"); ok {
		t.Error("expected false after Delete")
	}

	// 4. Clear
	c.Set("k1", "v1", 1*time.Second)
	c.Set("k2", "v2", 1*time.Second)
	c.Clear()
	if _, ok := c.Get("k1"); ok {
		t.Error("expected false after Clear for k1")
	}
	if _, ok := c.Get("k2"); ok {
		t.Error("expected false after Clear for k2")
	}
}

func TestMemoryCache_TTLLazyExpiration(t *testing.T) {
	c := NewMemoryCache(10 * time.Second) // 长间隔，验证惰性过期
	defer c.Close()

	c.Set("short_key", 12345, 50*time.Millisecond)

	// 立即获取有效
	val, ok := c.Get("short_key")
	if !ok || val != 12345 {
		t.Fatalf("expected 12345, got %v", val)
	}

	// 等待过期
	time.Sleep(60 * time.Millisecond)

	// 再次获取应被惰性删除
	if _, ok := c.Get("short_key"); ok {
		t.Error("expected short_key to be expired and lazily removed")
	}
}

func TestMemoryCache_PeriodicCleanup(t *testing.T) {
	c := NewMemoryCache(50 * time.Millisecond) // 短间隔周期清理
	defer c.Close()

	c.Set("exp1", "val", 20*time.Millisecond)
	c.Set("exp2", "val", 20*time.Millisecond)
	c.Set("keep", "val", 1*time.Second)

	// 等待后台 ticker 至少触发一次清理
	time.Sleep(100 * time.Millisecond)

	c.mu.RLock()
	count := len(c.items)
	c.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 active item after periodic cleanup, got %d", count)
	}
}

func TestMemoryCache_Concurrency(t *testing.T) {
	c := NewCache()
	defer c.Close()

	var wg sync.WaitGroup
	workers := 20
	ops := 100

	for i := range workers {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for j := range ops {
				k := fmt.Sprintf("k_%d_%d", w, j%5)
				c.Set(k, j, 500*time.Millisecond)
				c.Get(k)
				if j%10 == 0 {
					c.Delete(k)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestMemoryCache_CloseIdempotent(t *testing.T) {
	c := NewCache()
	c.Close()
	// 重复 Close 不应 panic
	c.Close()
}
