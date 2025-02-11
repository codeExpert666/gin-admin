// Package middleware 提供了一系列 Gin 框架的中间件工具函数
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SkippedPathPrefixes 检查当前请求路径是否以指定的前缀开始，用于判断是否需要跳过某些中间件
// 参数说明：
//   - c *gin.Context: Gin的上下文对象，包含了当前请求的所有信息
//   - prefixes: 可变参数，包含多个需要检查的路径前缀
//
// 返回值：
//   - bool: 如果当前路径以任意一个给定前缀开始，返回 true；否则返回 false
func SkippedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	// 如果没有提供前缀，则返回 false
	if len(prefixes) == 0 {
		return false
	}

	// 获取当前请求的路径
	path := c.Request.URL.Path
	pathLen := len(path)
	// 遍历所有前缀，检查当前路径是否以任意一个前缀开始
	for _, p := range prefixes {
		if pl := len(p); pathLen >= pl && path[:pl] == p {
			return true
		}
	}
	return false
}

// AllowedPathPrefixes 检查当前请求路径是否在允许的路径前缀列表中
// 参数说明：
//   - c *gin.Context: Gin的上下文对象
//   - prefixes: 可变参数，包含多个允许的路径前缀
//
// 返回值：
//   - bool: 如果没有指定前缀，返回 true；如果当前路径以任意一个允许的前缀开始，返回 true；否则返回 false
func AllowedPathPrefixes(c *gin.Context, prefixes ...string) bool {
	// 如果没有提供前缀，则默认允许所有路径
	if len(prefixes) == 0 {
		return true
	}

	// 获取当前请求的路径
	path := c.Request.URL.Path
	pathLen := len(path)
	// 遍历所有允许的前缀，检查当前路径是否匹配
	for _, p := range prefixes {
		if pl := len(p); pathLen >= pl && path[:pl] == p {
			return true
		}
	}
	return false
}

// Empty 返回一个空的中间件处理函数
// 这个中间件不做任何处理，直接调用下一个处理函数
// 常用于需要条件性地应用中间件时，作为默认的空操作处理器
func Empty() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
