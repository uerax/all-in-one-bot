package cache

import "time"

// Cache 定义通用并发安全缓存接口，与 store.Store 架构同级。
type Cache interface {
	// Get 获取指定 key 的缓存值，若不存在或已过期则返回 (nil, false)。
	Get(key string) (any, bool)
	// Set 设置缓存值并指定有效期 ttl。
	Set(key string, val any, ttl time.Duration)
	// Delete 删除指定 key 的缓存。
	Delete(key string)
	// Clear 清空所有缓存。
	Clear()
	// Close 关闭缓存后台资源（如定期清理 goroutine）。
	Close()
}

// NewCache 创建全局默认内存缓存实例（默认每分钟自动扫描清理一次过期条目）。
func NewCache() Cache {
	return NewMemoryCache(time.Minute)
}
