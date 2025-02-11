package cachex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig 定义 Redis 连接的配置结构
type RedisConfig struct {
	Addr     string // Redis 服务器地址，格式为 host:port
	Username string // Redis 用户名
	Password string // Redis 密码
	DB       int    // Redis 数据库编号
}

// NewRedisCache 创建一个基于 Redis 的缓存实例
// 参数:
// - cfg: Redis 配置信息
// - opts: 可选的配置选项
func NewRedisCache(cfg RedisConfig, opts ...Option) Cacher {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return newRedisCache(cli, opts...)
}

// NewRedisCacheWithClient 使用已存在的 Redis 客户端创建缓存实例
func NewRedisCacheWithClient(cli *redis.Client, opts ...Option) Cacher {
	return newRedisCache(cli, opts...)
}

// NewRedisCacheWithClusterClient 使用 Redis 集群客户端创建缓存实例
func NewRedisCacheWithClusterClient(cli *redis.ClusterClient, opts ...Option) Cacher {
	return newRedisCache(cli, opts...)
}

// newRedisCache 是创建 Redis 缓存的内部函数
// 参数:
// - cli: Redis 客户端接口
// - opts: 配置选项
func newRedisCache(cli redisClienter, opts ...Option) Cacher {
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	for _, o := range opts {
		o(defaultOpts)
	}

	return &redisCache{
		opts: defaultOpts,
		cli:  cli,
	}
}

// redisClienter 定义了 Redis 客户端需要实现的接口方法
type redisClienter interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Close() error
}

// redisCache 实现了 Cacher 接口的 Redis 缓存结构
type redisCache struct {
	opts *options      // 缓存配置选项
	cli  redisClienter // Redis 客户端
}

// getKey 生成带有命名空间的完整缓存键
func (a *redisCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, a.opts.Delimiter, key)
}

// Set 设置缓存键值对
// 参数:
// - ctx: 上下文
// - ns: 命名空间
// - key: 键名
// - value: 值
// - expiration: 可选的过期时间
func (a *redisCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	cmd := a.cli.Set(ctx, a.getKey(ns, key), value, exp)
	return cmd.Err()
}

// Get 获取缓存值
// 返回值:
// - string: 缓存的值
// - bool: 是否存在
// - error: 错误信息
func (a *redisCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	cmd := a.cli.Get(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return cmd.Val(), true, nil
}

// Exists 检查键是否存在
func (a *redisCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	cmd := a.cli.Exists(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Delete 删除缓存键
func (a *redisCache) Delete(ctx context.Context, ns, key string) error {
	b, err := a.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil
	}

	cmd := a.cli.Del(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return err
	}
	return nil
}

// GetAndDelete 获取缓存值并删除
// 这是一个原子操作，确保获取值后立即删除
func (a *redisCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := a.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	cmd := a.cli.Del(ctx, a.getKey(ns, key))
	if err := cmd.Err(); err != nil && err != redis.Nil {
		return "", false, err
	}
	return value, true, nil
}

// Iterator 遍历指定命名空间下的所有键值对
// 参数:
// - ctx: 上下文
// - ns: 命名空间
// - fn: 处理每个键值对的回调函数，返回 false 停止遍历
func (a *redisCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	var cursor uint64 = 0

LB_LOOP:
	for {
		// 使用 SCAN 命令分批获取键
		cmd := a.cli.Scan(ctx, cursor, a.getKey(ns, "*"), 100)
		if err := cmd.Err(); err != nil {
			return err
		}

		keys, c, err := cmd.Result()
		if err != nil {
			return err
		}

		// 遍历获取到的键
		for _, key := range keys {
			cmd := a.cli.Get(ctx, key)
			if err := cmd.Err(); err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}
			if next := fn(ctx, strings.TrimPrefix(key, a.getKey(ns, "")), cmd.Val()); !next {
				break LB_LOOP
			}
		}

		if c == 0 {
			break
		}
		cursor = c
	}

	return nil
}

// Close 关闭 Redis 连接
func (a *redisCache) Close(ctx context.Context) error {
	return a.cli.Close()
}
