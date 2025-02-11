// Package jwtx 提供了 JWT 令牌的存储和管理功能
package jwtx

import (
	"context"
	"time"
)

// Storer 定义了令牌存储的接口
// 该接口定义了存储、删除、检查和关闭等基本操作
type Storer interface {
	// Set 存储令牌，并设置过期时间
	Set(ctx context.Context, tokenStr string, expiration time.Duration) error
	// Delete 删除指定的令牌
	Delete(ctx context.Context, tokenStr string) error
	// Check 检查令牌是否存在
	Check(ctx context.Context, tokenStr string) (bool, error)
	// Close 关闭存储连接
	Close(ctx context.Context) error
}

// storeOptions 定义了存储的配置选项
type storeOptions struct {
	CacheNS string // 缓存的命名空间，默认值为 "jwt"
}

// StoreOption 是一个函数类型，用于设置存储选项
type StoreOption func(*storeOptions)

// WithCacheNS 返回一个设置缓存命名空间的选项函数
func WithCacheNS(ns string) StoreOption {
	return func(o *storeOptions) {
		o.CacheNS = ns
	}
}

// Cacher 定义了缓存操作的接口
// 这个接口提供了基本的缓存操作方法
type Cacher interface {
	// Set 设置缓存，可选过期时间
	Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error
	// Get 获取缓存的值
	Get(ctx context.Context, ns, key string) (string, bool, error)
	// Exists 检查键是否存在
	Exists(ctx context.Context, ns, key string) (bool, error)
	// Delete 删除缓存的键
	Delete(ctx context.Context, ns, key string) error
	// Close 关闭缓存连接
	Close(ctx context.Context) error
}

// NewStoreWithCache 创建一个基于缓存的存储实现
// 参数 cache 是实现了 Cacher 接口的缓存对象
// 参数 opts 是可选的配置选项
func NewStoreWithCache(cache Cacher, opts ...StoreOption) Storer {
	s := &storeImpl{
		c: cache,
		opts: &storeOptions{
			CacheNS: "jwt", // 设置默认的命名空间
		},
	}
	// 应用所有的配置选项
	for _, opt := range opts {
		opt(s.opts)
	}
	return s
}

// storeImpl 是 Storer 接口的具体实现
type storeImpl struct {
	opts *storeOptions // 存储配置选项
	c    Cacher        // 缓存实现
}

// Set 实现了 Storer 接口的 Set 方法
// 将令牌存储到缓存中，并设置过期时间
func (s *storeImpl) Set(ctx context.Context, tokenStr string, expiration time.Duration) error {
	return s.c.Set(ctx, s.opts.CacheNS, tokenStr, "", expiration)
}

// Delete 实现了 Storer 接口的 Delete 方法
// 从缓存中删除指定的令牌
func (s *storeImpl) Delete(ctx context.Context, tokenStr string) error {
	return s.c.Delete(ctx, s.opts.CacheNS, tokenStr)
}

// Check 实现了 Storer 接口的 Check 方法
// 检查令牌是否存在于缓存中
func (s *storeImpl) Check(ctx context.Context, tokenStr string) (bool, error) {
	return s.c.Exists(ctx, s.opts.CacheNS, tokenStr)
}

// Close 实现了 Storer 接口的 Close 方法
// 关闭缓存连接
func (s *storeImpl) Close(ctx context.Context) error {
	return s.c.Close(ctx)
}
