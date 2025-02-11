// Package middleware 提供了 Web 中间件相关功能
package middleware

import (
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// ErrCasbinDenied 定义了权限被拒绝时的错误
// 使用 errors.Forbidden 创建一个 HTTP 403 禁止访问的错误
var ErrCasbinDenied = errors.Forbidden("com.casbin.denied", "Permission denied")

// CasbinConfig 定义了 Casbin 中间件的配置结构
type CasbinConfig struct {
	// AllowedPathPrefixes 定义允许访问的路径前缀列表
	AllowedPathPrefixes []string
	// SkippedPathPrefixes 定义跳过权限检查的路径前缀列表
	SkippedPathPrefixes []string
	// Skipper 定义了一个函数，用于判断是否跳过当前请求的权限检查
	Skipper func(c *gin.Context) bool
	// GetEnforcer 定义了获取 Casbin enforcer 实例的函数
	GetEnforcer func(c *gin.Context) *casbin.Enforcer
	// GetSubjects 定义了获取当前请求主体（通常是用户角色）的函数
	GetSubjects func(c *gin.Context) []string
}

// CasbinWithConfig 创建一个基于配置的 Casbin 中间件
// 返回一个 Gin 中间件处理函数
func CasbinWithConfig(config CasbinConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过权限验证：
		// 1. 如果请求路径不在允许的前缀列表中
		// 2. 如果请求路径在跳过的前缀列表中
		// 3. 如果自定义的 Skipper 函数返回 true
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			(config.Skipper != nil && config.Skipper(c)) {
			c.Next() // 跳过权限检查，继续处理请求
			return
		}

		// 获取 Casbin enforcer 实例
		enforcer := config.GetEnforcer(c)
		if enforcer == nil {
			util.ResError(c, ErrCasbinDenied)
			return
		}

		// 遍历所有主体（subjects），检查是否有任一主体具有访问权限
		for _, sub := range config.GetSubjects(c) {
			// 使用 Casbin enforcer 检查权限
			// 参数：主体、请求路径、请求方法
			if b, err := enforcer.Enforce(sub, c.Request.URL.Path, c.Request.Method); err != nil {
				util.ResError(c, err)
				return
			} else if b {
				// 如果有权限，继续处理请求
				c.Next()
				return
			}
		}
		// 如果所有主体都没有权限，返回权限拒绝错误
		util.ResError(c, ErrCasbinDenied)
	}
}
