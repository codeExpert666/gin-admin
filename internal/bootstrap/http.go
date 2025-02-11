package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/utility/prom"
	"github.com/LyricTian/gin-admin/v10/internal/wirex"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/middleware"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// startHTTPServer 启动 HTTP 服务器
// 参数:
// - ctx: 上下文，用于控制服务器的生命周期
// - injector: 依赖注入器，包含所有需要的服务实例
// 返回:
// - 清理函数：用于优雅关闭服务器
// - 错误信息
func startHTTPServer(ctx context.Context, injector *wirex.Injector) (func(), error) {
	// 根据配置设置 Gin 的运行模式（开发/发布）
	if config.C.IsDebug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 引擎实例
	e := gin.New()

	// 添加健康检查接口
	e.GET("/health", func(c *gin.Context) {
		util.ResOK(c)
	})

	// 配置全局异常恢复中间件
	e.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
		Skip: config.C.Middleware.Recovery.Skip,
	}))

	// 处理不支持的 HTTP 方法
	e.NoMethod(func(c *gin.Context) {
		util.ResError(c, errors.MethodNotAllowed("", "Method Not Allowed"))
	})

	// 处理未找到的路由
	e.NoRoute(func(c *gin.Context) {
		util.ResError(c, errors.NotFound("", "Not Found"))
	})

	// 获取允许的路由前缀
	allowedPrefixes := injector.M.RouterPrefixes()

	// 注册中间件
	if err := useHTTPMiddlewares(ctx, e, injector, allowedPrefixes); err != nil {
		return nil, err
	}

	// 注册业务路由
	if err := injector.M.RegisterRouters(ctx, e); err != nil {
		return nil, err
	}

	// 配置 Swagger 文档
	if !config.C.General.DisableSwagger {
		e.StaticFile("/openapi.json", filepath.Join(config.C.General.WorkDir, "openapi.json"))
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 配置静态文件服务
	if dir := config.C.Middleware.Static.Dir; dir != "" {
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:                dir,
			SkippedPathPrefixes: allowedPrefixes,
		}))
	}

	// 配置 HTTP 服务器
	addr := config.C.General.HTTP.Addr
	logging.Context(ctx).Info(fmt.Sprintf("HTTP server is listening on %s", addr))
	srv := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  time.Second * time.Duration(config.C.General.HTTP.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(config.C.General.HTTP.WriteTimeout),
		IdleTimeout:  time.Second * time.Duration(config.C.General.HTTP.IdleTimeout),
	}

	// 在后台启动服务器
	go func() {
		var err error
		// 判断是否启用 HTTPS
		if config.C.General.HTTP.CertFile != "" && config.C.General.HTTP.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(config.C.General.HTTP.CertFile, config.C.General.HTTP.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logging.Context(ctx).Error("Failed to listen http server", zap.Error(err))
		}
	}()

	// 返回清理函数，用于优雅关闭服务器
	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(config.C.General.HTTP.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logging.Context(ctx).Error("Failed to shutdown http server", zap.Error(err))
		}
	}, nil
}

// useHTTPMiddlewares 配置 HTTP 中间件
// 参数:
// - ctx: 上下文
// - e: Gin 引擎实例
// - injector: 依赖注入器
// - allowedPrefixes: 允许的路由前缀列表
func useHTTPMiddlewares(_ context.Context, e *gin.Engine, injector *wirex.Injector, allowedPrefixes []string) error {
	// 配置 CORS 中间件（跨域资源共享）
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Enable:                 config.C.Middleware.CORS.Enable,
		AllowAllOrigins:        config.C.Middleware.CORS.AllowAllOrigins,
		AllowOrigins:           config.C.Middleware.CORS.AllowOrigins,
		AllowMethods:           config.C.Middleware.CORS.AllowMethods,
		AllowHeaders:           config.C.Middleware.CORS.AllowHeaders,
		AllowCredentials:       config.C.Middleware.CORS.AllowCredentials,
		ExposeHeaders:          config.C.Middleware.CORS.ExposeHeaders,
		MaxAge:                 config.C.Middleware.CORS.MaxAge,
		AllowWildcard:          config.C.Middleware.CORS.AllowWildcard,
		AllowBrowserExtensions: config.C.Middleware.CORS.AllowBrowserExtensions,
		AllowWebSockets:        config.C.Middleware.CORS.AllowWebSockets,
		AllowFiles:             config.C.Middleware.CORS.AllowFiles,
	}))

	// 配置请求追踪中间件
	e.Use(middleware.TraceWithConfig(middleware.TraceConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Trace.SkippedPathPrefixes,
		RequestHeaderKey:    config.C.Middleware.Trace.RequestHeaderKey,
		ResponseTraceKey:    config.C.Middleware.Trace.ResponseTraceKey,
	}))

	// 配置日志中间件
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		AllowedPathPrefixes:      allowedPrefixes,
		SkippedPathPrefixes:      config.C.Middleware.Logger.SkippedPathPrefixes,
		MaxOutputRequestBodyLen:  config.C.Middleware.Logger.MaxOutputRequestBodyLen,
		MaxOutputResponseBodyLen: config.C.Middleware.Logger.MaxOutputResponseBodyLen,
	}))

	// 配置请求体复制中间件（用于日志记录等）
	e.Use(middleware.CopyBodyWithConfig(middleware.CopyBodyConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.CopyBody.SkippedPathPrefixes,
		MaxContentLen:       config.C.Middleware.CopyBody.MaxContentLen,
	}))

	// 配置认证中间件
	e.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Auth.SkippedPathPrefixes,
		ParseUserID:         injector.M.RBAC.LoginAPI.LoginBIZ.ParseUserID,
		RootID:              config.C.General.Root.ID,
	}))

	// 配置限流中间件
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Enable:              config.C.Middleware.RateLimiter.Enable,
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.RateLimiter.SkippedPathPrefixes,
		Period:              config.C.Middleware.RateLimiter.Period,
		MaxRequestsPerIP:    config.C.Middleware.RateLimiter.MaxRequestsPerIP,
		MaxRequestsPerUser:  config.C.Middleware.RateLimiter.MaxRequestsPerUser,
		StoreType:           config.C.Middleware.RateLimiter.Store.Type,
		MemoryStoreConfig: middleware.RateLimiterMemoryConfig{
			Expiration:      time.Second * time.Duration(config.C.Middleware.RateLimiter.Store.Memory.Expiration),
			CleanupInterval: time.Second * time.Duration(config.C.Middleware.RateLimiter.Store.Memory.CleanupInterval),
		},
		RedisStoreConfig: middleware.RateLimiterRedisConfig{
			Addr:     config.C.Middleware.RateLimiter.Store.Redis.Addr,
			Password: config.C.Middleware.RateLimiter.Store.Redis.Password,
			DB:       config.C.Middleware.RateLimiter.Store.Redis.DB,
			Username: config.C.Middleware.RateLimiter.Store.Redis.Username,
		},
	}))

	// 配置访问控制中间件（Casbin）
	e.Use(middleware.CasbinWithConfig(middleware.CasbinConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: config.C.Middleware.Casbin.SkippedPathPrefixes,
		Skipper: func(c *gin.Context) bool {
			// 如果禁用了 Casbin 或者是 Root 用户，则跳过权限检查
			if config.C.Middleware.Casbin.Disable ||
				util.FromIsRootUser(c.Request.Context()) {
				return true
			}
			return false
		},
		GetEnforcer: func(c *gin.Context) *casbin.Enforcer {
			return injector.M.RBAC.Casbinx.GetEnforcer()
		},
		GetSubjects: func(c *gin.Context) []string {
			return util.FromUserCache(c.Request.Context()).RoleIDs
		},
	}))

	// 配置 Prometheus 监控中间件
	if config.C.Util.Prometheus.Enable {
		e.Use(prom.GinMiddleware)
	}

	return nil
}
