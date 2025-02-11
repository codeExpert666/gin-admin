// Package middleware 提供了各种 Gin 框架的中间件
package middleware

import (
	"fmt"
	"strings"

	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid" // 用于生成唯一ID
)

// TraceConfig 定义了跟踪中间件的配置选项
type TraceConfig struct {
	AllowedPathPrefixes []string // 允许进行跟踪的路径前缀列表
	SkippedPathPrefixes []string // 需要跳过跟踪的路径前缀列表
	RequestHeaderKey    string   // 请求头中跟踪ID的键名
	ResponseTraceKey    string   // 响应头中跟踪ID的键名
}

// DefaultTraceConfig 定义了默认的跟踪配置
var DefaultTraceConfig = TraceConfig{
	RequestHeaderKey: "X-Request-Id", // 默认请求头跟踪ID键名
	ResponseTraceKey: "X-Trace-Id",   // 默认响应头跟踪ID键名
}

// Trace 使用默认配置创建一个跟踪中间件
func Trace() gin.HandlerFunc {
	return TraceWithConfig(DefaultTraceConfig)
}

// TraceWithConfig 根据自定义配置创建跟踪中间件
// 该中间件的主要功能是为每个请求生成或使用跟踪ID，用于请求追踪和日志记录
func TraceWithConfig(config TraceConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求路径是否满足跟踪条件
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 尝试从请求头获取跟踪ID，如果没有则生成新的跟踪ID
		traceID := c.GetHeader(config.RequestHeaderKey)
		if traceID == "" {
			// 使用 xid 生成唯一的跟踪ID，并添加 "TRACE-" 前缀
			traceID = fmt.Sprintf("TRACE-%s", strings.ToUpper(xid.New().String()))
		}

		// 将跟踪ID添加到请求上下文中
		ctx := util.NewTraceID(c.Request.Context(), traceID)
		ctx = logging.NewTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 在响应头中设置跟踪ID
		c.Writer.Header().Set(config.ResponseTraceKey, traceID)

		// 继续处理后续的中间件和路由处理函数
		c.Next()
	}
}
