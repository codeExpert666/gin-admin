// Package jwtx 提供了 JWT 相关的功能实现
package jwtx

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// defaultDelimiter 定义了缓存键的分隔符
var defaultDelimiter = ":"

// MemoryConfig 内存缓存的配置结构
type MemoryConfig struct {
	// CleanupInterval 定义了清理过期缓存项的时间间隔
	CleanupInterval time.Duration
}

// NewMemoryCache 创建一个新的内存缓存实例
// 参数 cfg 用于配置缓存的清理间隔时间
func NewMemoryCache(cfg MemoryConfig) Cacher {
	return &memCache{
		cache: cache.New(0, cfg.CleanupInterval),
	}
}

// memCache 实现了基于内存的缓存结构
type memCache struct {
	cache *cache.Cache // 使用 go-cache 库实现的内存缓存
}

// getKey 生成缓存的键名
// ns 为命名空间，key 为具体的键名
// 返回格式为: "namespace:key"
func (a *memCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, defaultDelimiter, key)
}

// Set 设置缓存值
// ctx: 上下文信息
// ns: 命名空间
// key: 键名
// value: 要存储的值
// expiration: 可选的过期时间，如果提供则设置过期时间
func (a *memCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	a.cache.Set(a.getKey(ns, key), value, exp)
	return nil
}

// Get 获取缓存值
// 返回值：
// - string: 缓存的值
// - bool: 是否存在该键
// - error: 错误信息
func (a *memCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	val, ok := a.cache.Get(a.getKey(ns, key))
	if !ok {
		return "", false, nil
	}
	return val.(string), ok, nil
}

// Exists 检查键是否存在
// 返回 true 表示键存在，false 表示键不存在
func (a *memCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	_, ok := a.cache.Get(a.getKey(ns, key))
	return ok, nil
}

// Delete 删除指定的缓存项
func (a *memCache) Delete(ctx context.Context, ns, key string) error {
	a.cache.Delete(a.getKey(ns, key))
	return nil
}

// Close 关闭缓存，清空所有缓存项
func (a *memCache) Close(ctx context.Context) error {
	a.cache.Flush()
	return nil
}
