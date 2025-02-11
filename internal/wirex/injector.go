// Package wirex 实现了依赖注入的功能，用于初始化和管理应用程序的核心组件
package wirex

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/gormx"
	"github.com/LyricTian/gin-admin/v10/pkg/jwtx"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// Injector 结构体定义了应用程序的核心依赖组件
type Injector struct {
	DB    *gorm.DB      // 数据库连接实例
	Cache cachex.Cacher // 缓存服务实例
	Auth  jwtx.Auther   // JWT认证服务实例
	M     *mods.Mods    // 模块管理器实例
}

// InitDB 初始化数据库连接
// 参数：
// - ctx: 上下文对象，用于控制数据库操作的生命周期
// 返回值：
// - *gorm.DB: 数据库连接实例
// - func(): 清理函数，用于关闭数据库连接
// - error: 可能发生的错误
func InitDB(ctx context.Context) (*gorm.DB, func(), error) {
	// 获取数据库配置
	cfg := config.C.Storage.DB

	// 构建数据库解析器配置
	resolver := make([]gormx.ResolverConfig, len(cfg.Resolver))
	for i, v := range cfg.Resolver {
		resolver[i] = gormx.ResolverConfig{
			DBType:   v.DBType,   // 数据库类型
			Sources:  v.Sources,  // 主数据库源
			Replicas: v.Replicas, // 从数据库源
			Tables:   v.Tables,   // 相关的数据表
		}
	}

	// 创建数据库连接
	db, err := gormx.New(gormx.Config{
		Debug:        cfg.Debug,        // 是否开启调试模式
		PrepareStmt:  cfg.PrepareStmt,  // 是否启用预编译语句
		DBType:       cfg.Type,         // 数据库类型
		DSN:          cfg.DSN,          // 数据库连接字符串
		MaxLifetime:  cfg.MaxLifetime,  // 连接最大生命周期
		MaxIdleTime:  cfg.MaxIdleTime,  // 空闲连接最大生命周期
		MaxOpenConns: cfg.MaxOpenConns, // 最大打开连接数
		MaxIdleConns: cfg.MaxIdleConns, // 最大空闲连接数
		TablePrefix:  cfg.TablePrefix,  // 数据表前缀
		Resolver:     resolver,         // 数据库解析器配置
	})
	if err != nil {
		return nil, nil, err
	}

	// 返回数据库实例和清理函数
	return db, func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}, nil
}

// InitCacher 初始化缓存服务
// 参数：
// - ctx: 上下文对象，用于控制缓存操作的生命周期
// 返回值：
// - cachex.Cacher: 缓存服务实例
// - func(): 清理函数，用于关闭缓存连接
// - error: 可能发生的错误
func InitCacher(ctx context.Context) (cachex.Cacher, func(), error) {
	// 获取缓存配置
	cfg := config.C.Storage.Cache

	var cache cachex.Cacher
	// 根据配置类型选择不同的缓存实现
	switch cfg.Type {
	case "redis": // Redis缓存
		cache = cachex.NewRedisCache(cachex.RedisConfig{
			Addr:     cfg.Redis.Addr,
			DB:       cfg.Redis.DB,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
		}, cachex.WithDelimiter(cfg.Delimiter))
	case "badger": // Badger缓存（基于磁盘的KV存储）
		cache = cachex.NewBadgerCache(cachex.BadgerConfig{
			Path: cfg.Badger.Path,
		}, cachex.WithDelimiter(cfg.Delimiter))
	default: // 默认使用内存缓存
		cache = cachex.NewMemoryCache(cachex.MemoryConfig{
			CleanupInterval: time.Second * time.Duration(cfg.Memory.CleanupInterval),
		}, cachex.WithDelimiter(cfg.Delimiter))
	}

	return cache, func() {
		_ = cache.Close(ctx)
	}, nil
}

// InitAuth 初始化JWT认证服务
// 参数：
// - ctx: 上下文对象，用于控制认证服务的生命周期
// 返回值：
// - jwtx.Auther: JWT认证服务实例
// - func(): 清理函数，用于释放认证服务资源
// - error: 可能发生的错误
func InitAuth(ctx context.Context) (jwtx.Auther, func(), error) {
	// 获取认证配置
	cfg := config.C.Middleware.Auth
	var opts []jwtx.Option
	// 设置JWT选项
	opts = append(opts, jwtx.SetExpired(cfg.Expired))
	opts = append(opts, jwtx.SetSigningKey(cfg.SigningKey, cfg.OldSigningKey))

	// 选择JWT签名方法
	var method jwt.SigningMethod
	switch cfg.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	default:
		method = jwt.SigningMethodHS512
	}
	opts = append(opts, jwtx.SetSigningMethod(method))

	// 初始化认证服务的缓存存储
	var cache cachex.Cacher
	switch cfg.Store.Type {
	case "redis":
		cache = cachex.NewRedisCache(cachex.RedisConfig{
			Addr:     cfg.Store.Redis.Addr,
			DB:       cfg.Store.Redis.DB,
			Username: cfg.Store.Redis.Username,
			Password: cfg.Store.Redis.Password,
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	case "badger":
		cache = cachex.NewBadgerCache(cachex.BadgerConfig{
			Path: cfg.Store.Badger.Path,
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	default:
		cache = cachex.NewMemoryCache(cachex.MemoryConfig{
			CleanupInterval: time.Second * time.Duration(cfg.Store.Memory.CleanupInterval),
		}, cachex.WithDelimiter(cfg.Store.Delimiter))
	}

	// 创建认证服务实例
	auth := jwtx.New(jwtx.NewStoreWithCache(cache), opts...)
	return auth, func() {
		_ = auth.Release(ctx)
	}, nil
}
