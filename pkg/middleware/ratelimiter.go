// Package middleware 实现了各种 Gin 框架的中间件
package middleware

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiterConfig 定义了限流器的配置结构
type RateLimiterConfig struct {
	Enable              bool                    // 是否启用限流器
	AllowedPathPrefixes []string                // 允许进行限流的路径前缀列表
	SkippedPathPrefixes []string                // 跳过限流的路径前缀列表
	Period              int                     // 限流周期（秒）
	MaxRequestsPerIP    int                     // 每个IP在周期内的最大请求数
	MaxRequestsPerUser  int                     // 每个用户在周期内的最大请求数
	StoreType           string                  // 存储类型：memory（内存）或 redis
	MemoryStoreConfig   RateLimiterMemoryConfig // 内存存储配置
	RedisStoreConfig    RateLimiterRedisConfig  // Redis存储配置
}

// RateLimiterWithConfig 创建一个基于配置的限流中间件
func RateLimiterWithConfig(config RateLimiterConfig) gin.HandlerFunc {
	// 如果限流器未启用，返回空中间件
	if !config.Enable {
		return Empty()
	}

	// 根据配置选择存储类型
	var store RateLimiterStorer
	switch config.StoreType {
	case "redis":
		store = NewRateLimiterRedisStore(config.RedisStoreConfig)
	default:
		store = NewRateLimiterMemoryStore(config.MemoryStoreConfig)
	}

	// 返回限流中间件处理函数
	return func(c *gin.Context) {
		// 检查请求路径是否需要进行限流
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		var (
			allowed bool
			err     error
		)

		ctx := c.Request.Context()
		// 优先使用用户ID进行限流，如果没有用户ID则使用IP地址
		if userID := util.FromUserID(ctx); userID != "" {
			allowed, err = store.Allow(ctx, userID, time.Second*time.Duration(config.Period), config.MaxRequestsPerUser)
		} else {
			allowed, err = store.Allow(ctx, c.ClientIP(), time.Second*time.Duration(config.Period), config.MaxRequestsPerIP)
		}

		// 处理限流结果
		if err != nil {
			logging.Context(ctx).Error("限流器中间件错误", zap.Error(err))
			util.ResError(c, errors.InternalServerError("", "服务器内部错误，请稍后重试。"))
		} else if allowed {
			c.Next()
		} else {
			util.ResError(c, errors.TooManyRequests("", "请求过于频繁，请稍后重试。"))
		}
	}
}

// RateLimiterStorer 定义了限流器存储接口
type RateLimiterStorer interface {
	// Allow 检查请求是否允许通过
	Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error)
}

// 内存存储相关实现
// NewRateLimiterMemoryStore 创建基于内存的限流器存储
func NewRateLimiterMemoryStore(config RateLimiterMemoryConfig) RateLimiterStorer {
	return &RateLimiterMemoryStore{
		cache: cache.New(config.Expiration, config.CleanupInterval),
	}
}

// RateLimiterMemoryConfig 定义内存存储的配置
type RateLimiterMemoryConfig struct {
	Expiration      time.Duration // 过期时间
	CleanupInterval time.Duration // 清理间隔
}

// RateLimiterMemoryStore 实现了基于内存的限流器存储
type RateLimiterMemoryStore struct {
	cache *cache.Cache
}

// Allow 实现了基于内存的限流检查
func (s *RateLimiterMemoryStore) Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error) {
	// 如果周期或最大请求数无效，直接允许请求
	if period.Seconds() <= 0 || maxRequests <= 0 {
		return true, nil
	}

	// 检查是否存在现有的限流器
	if limiter, exists := s.cache.Get(identifier); exists {
		isAllow := limiter.(*rate.Limiter).Allow()
		s.cache.SetDefault(identifier, limiter)
		return isAllow, nil
	}

	// 创建新的限流器
	limiter := rate.NewLimiter(rate.Every(period), maxRequests)
	limiter.Allow()
	s.cache.SetDefault(identifier, limiter)

	return true, nil
}

// Redis存储相关实现
// RateLimiterRedisConfig 定义Redis存储的配置
type RateLimiterRedisConfig struct {
	Addr     string // Redis服务器地址
	Username string // Redis用户名
	Password string // Redis密码
	DB       int    // Redis数据库编号
}

// NewRateLimiterRedisStore 创建基于Redis的限流器存储
func NewRateLimiterRedisStore(config RateLimiterRedisConfig) RateLimiterStorer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RateLimiterRedisStore{
		limiter: redis_rate.NewLimiter(rdb),
	}
}

// RateLimiterRedisStore 实现了基于Redis的限流器存储
type RateLimiterRedisStore struct {
	limiter *redis_rate.Limiter
}

// Allow 实现了基于Redis的限流检查
func (s *RateLimiterRedisStore) Allow(ctx context.Context, identifier string, period time.Duration, maxRequests int) (bool, error) {
	// 如果周期或最大请求数无效，直接允许请求
	if period.Seconds() <= 0 || maxRequests <= 0 {
		return true, nil
	}

	// 使用Redis进行限流检查
	result, err := s.limiter.Allow(ctx, identifier, redis_rate.PerSecond(maxRequests/int(period.Seconds())))
	if err != nil {
		return false, err
	}
	return result.Allowed > 0, nil
}
