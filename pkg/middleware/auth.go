// Package middleware 提供了 Web 应用的中间件功能
package middleware

import (
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
)

// AuthConfig 定义了认证中间件的配置结构
type AuthConfig struct {
	// AllowedPathPrefixes 定义允许访问的路径前缀列表
	AllowedPathPrefixes []string
	// SkippedPathPrefixes 定义跳过认证的路径前缀列表
	SkippedPathPrefixes []string
	// RootID 定义超级管理员的用户ID
	RootID string
	// Skipper 是一个自定义函数，用于判断是否跳过当前请求的认证
	// 返回 true 时跳过认证
	Skipper func(c *gin.Context) bool
	// ParseUserID 是一个自定义函数，用于从请求中解析用户ID
	// 通常从 JWT token 或者其他认证信息中获取
	ParseUserID func(c *gin.Context) (string, error)
}

// AuthWithConfig 创建一个基于给定配置的认证中间件
// 返回一个 Gin 中间件处理函数
func AuthWithConfig(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过认证：
		// 1. 如果请求路径不在允许的前缀列表中
		// 2. 如果请求路径在跳过认证的前缀列表中
		// 3. 如果自定义的 Skipper 函数返回 true
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next()
			return
		}

		// 解析用户ID
		userID, err := config.ParseUserID(c)
		if err != nil {
			// 如果解析失败，返回错误响应
			util.ResError(c, err)
			return
		}

		// 创建新的上下文，包含用户信息：
		// 1. 将用户ID添加到请求上下文中
		ctx := util.NewUserID(c.Request.Context(), userID)
		// 2. 将用户ID添加到日志上下文中
		ctx = logging.NewUserID(ctx, userID)
		// 3. 如果是超级管理员，添加root用户标识
		if userID == config.RootID {
			ctx = util.NewIsRootUser(ctx)
		}
		// 使用新的上下文继续处理请求
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
