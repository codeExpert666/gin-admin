package cachex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cacher 定义了缓存操作的基本接口
// 包含了对缓存进行设置、获取和删除等基本操作的方法
type Cacher interface {
	// Set 设置缓存值，可选过期时间
	// ns: 命名空间，用于隔离不同的缓存域
	// key: 缓存键
	// value: 缓存值
	// expiration: 可选的过期时间
	Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error

	// Get 获取缓存值
	// 返回值: (缓存的值, 是否存在, 错误信息)
	Get(ctx context.Context, ns, key string) (string, bool, error)

	// GetAndDelete 获取缓存值并删除
	// 返回值: (缓存的值, 是否存在, 错误信息)
	GetAndDelete(ctx context.Context, ns, key string) (string, bool, error)

	// Exists 检查缓存键是否存在
	Exists(ctx context.Context, ns, key string) (bool, error)

	// Delete 删除指定的缓存项
	Delete(ctx context.Context, ns, key string) error

	// Iterator 遍历指定命名空间下的所有缓存项
	// fn 为遍历回调函数，返回 false 时停止遍历
	Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error

	// Close 关闭缓存，清理资源
	Close(ctx context.Context) error
}

// 默认的分隔符，用于连接命名空间和键名
var defaultDelimiter = ":"

// options 定义缓存配置选项
type options struct {
	Delimiter string // 分隔符
}

// Option 定义配置函数类型
type Option func(*options)

// WithDelimiter 设置自定义分隔符的配置函数
func WithDelimiter(delimiter string) Option {
	return func(o *options) {
		o.Delimiter = delimiter
	}
}

// MemoryConfig 内存缓存的配置结构
type MemoryConfig struct {
	CleanupInterval time.Duration // 清理过期项的时间间隔
}

// NewMemoryCache 创建一个新的内存缓存实例
// cfg: 缓存配置
// opts: 可选的配置选项
func NewMemoryCache(cfg MemoryConfig, opts ...Option) Cacher {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}

	return &memCache{
		opts:  defaultOpts,
		cache: cache.New(0, cfg.CleanupInterval),
	}
}

// memCache 实现了基于内存的缓存
type memCache struct {
	opts  *options     // 配置选项
	cache *cache.Cache // 底层使用 go-cache 实现
}

// getKey 生成完整的缓存键
// 通过将命名空间和键名用分隔符连接
func (a *memCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, a.opts.Delimiter, key)
}

// Set 实现缓存值的设置
func (a *memCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	a.cache.Set(a.getKey(ns, key), value, exp)
	return nil
}

// Get 实现缓存值的获取
func (a *memCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	val, ok := a.cache.Get(a.getKey(ns, key))
	if !ok {
		return "", false, nil
	}
	return val.(string), ok, nil
}

// Exists 实现缓存键存在性检查
func (a *memCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	_, ok := a.cache.Get(a.getKey(ns, key))
	return ok, nil
}

// Delete 实现缓存项的删除
func (a *memCache) Delete(ctx context.Context, ns, key string) error {
	a.cache.Delete(a.getKey(ns, key))
	return nil
}

// GetAndDelete 实现获取并删除缓存项
func (a *memCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := a.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	a.cache.Delete(a.getKey(ns, key))
	return value, true, nil
}

// Iterator 实现缓存项的遍历
// 遍历指定命名空间下的所有缓存项
func (a *memCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	for k, v := range a.cache.Items() {
		if strings.HasPrefix(k, a.getKey(ns, "")) {
			if !fn(ctx, strings.TrimPrefix(k, a.getKey(ns, "")), v.Object.(string)) {
				break
			}
		}
	}
	return nil
}

// Close 实现缓存的关闭
// 清空所有缓存项
func (a *memCache) Close(ctx context.Context) error {
	a.cache.Flush()
	return nil
}
