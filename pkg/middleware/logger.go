// Package middleware 提供了各种 Gin 框架的中间件
package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerConfig 定义了日志中间件的配置选项
type LoggerConfig struct {
	AllowedPathPrefixes      []string // 允许记录日志的路径前缀列表
	SkippedPathPrefixes      []string // 跳过记录日志的路径前缀列表
	MaxOutputRequestBodyLen  int      // 记录请求体的最大长度（单位：字节）
	MaxOutputResponseBodyLen int      // 记录响应体的最大长度（单位：字节）
}

// DefaultLoggerConfig 定义了默认的日志中间件配置
var DefaultLoggerConfig = LoggerConfig{
	MaxOutputRequestBodyLen:  1024 * 1024, // 默认最大请求体记录长度为1MB
	MaxOutputResponseBodyLen: 1024 * 1024, // 默认最大响应体记录长度为1MB
}

// Logger 创建一个使用默认配置的日志中间件
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig 根据自定义配置创建日志中间件
func LoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求路径是否需要记录日志
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 记录请求开始时间
		start := time.Now()
		contentType := c.Request.Header.Get("Content-Type")

		// 构建基础日志字段
		fields := []zap.Field{
			zap.String("client_ip", c.ClientIP()),                // 客户端IP
			zap.String("method", c.Request.Method),               // HTTP方法
			zap.String("path", c.Request.URL.Path),               // 请求路径
			zap.String("user_agent", c.Request.UserAgent()),      // 用户代理
			zap.String("referer", c.Request.Referer()),           // 请求来源
			zap.String("uri", c.Request.RequestURI),              // 完整URI
			zap.String("host", c.Request.Host),                   // 主机名
			zap.String("remote_addr", c.Request.RemoteAddr),      // 远程地址
			zap.String("proto", c.Request.Proto),                 // HTTP协议版本
			zap.Int64("content_length", c.Request.ContentLength), // 内容长度
			zap.String("content_type", contentType),              // 内容类型
			zap.String("pragma", c.Request.Header.Get("Pragma")), // Pragma头部
		}

		// 继续处理请求
		c.Next()

		// 对于POST或PUT请求，记录请求体内容（如果是JSON格式）
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			mediaType, _, _ := mime.ParseMediaType(contentType)
			if mediaType == "application/json" {
				if v, ok := c.Get(util.ReqBodyKey); ok {
					if b, ok := v.([]byte); ok && len(b) <= config.MaxOutputRequestBodyLen {
						fields = append(fields, zap.String("body", string(b)))
					}
				}
			}
		}

		// 计算请求处理时间并添加到日志字段
		cost := time.Since(start).Nanoseconds() / 1e6
		fields = append(fields, zap.Int64("cost", cost))                                              // 处理耗时（毫秒）
		fields = append(fields, zap.Int("status", c.Writer.Status()))                                 // HTTP状态码
		fields = append(fields, zap.String("res_time", time.Now().Format("2006-01-02 15:04:05.999"))) // 响应时间
		fields = append(fields, zap.Int("res_size", c.Writer.Size()))                                 // 响应大小

		// 记录响应体内容（如果在最大长度限制内）
		if v, ok := c.Get(util.ResBodyKey); ok {
			if b, ok := v.([]byte); ok && len(b) <= config.MaxOutputResponseBodyLen {
				fields = append(fields, zap.String("res_body", string(b)))
			}
		}

		// 创建带有请求标签的上下文并记录日志
		ctx := c.Request.Context()
		ctx = logging.NewTag(ctx, logging.TagKeyRequest)
		logging.Context(ctx).Info(fmt.Sprintf("[HTTP] %s-%s-%d (%dms)",
			c.Request.URL.Path, c.Request.Method, c.Writer.Status(), cost), fields...)
	}
}
