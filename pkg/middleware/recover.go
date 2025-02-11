// Package middleware 提供了各种 Gin 框架的中间件
package middleware

import (
	"fmt"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryConfig 定义了恢复中间件的配置结构
type RecoveryConfig struct {
	Skip int // 跳过的堆栈帧数，默认为3
}

// DefaultRecoveryConfig 定义了默认的恢复中间件配置
var DefaultRecoveryConfig = RecoveryConfig{
	Skip: 3,
}

// Recovery 使用默认配置创建一个恢复中间件
// 该中间件用于捕获任何 panic，并返回 500 状态码
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig)
}

// RecoveryWithConfig 根据自定义配置创建恢复中间件
// 这个中间件的主要作用是：
// 1. 捕获程序运行时的 panic
// 2. 记录详细的错误日志
// 3. 返回友好的错误响应给客户端
func RecoveryWithConfig(config RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 defer 确保在程序 panic 时能够被捕获
		defer func() {
			// 尝试恢复 panic
			if rv := recover(); rv != nil {
				// 获取请求上下文并添加恢复标签
				ctx := c.Request.Context()
				ctx = logging.NewTag(ctx, logging.TagKeyRecovery)

				// 构建错误日志字段
				var fields []zap.Field
				// 添加错误信息
				fields = append(fields, zap.Strings("error", []string{fmt.Sprintf("%v", rv)}))
				// 添加堆栈信息
				fields = append(fields, zap.StackSkip("stack", config.Skip))

				// 在调试模式下，添加额外的请求头信息
				if gin.IsDebugging() {
					// 获取请求信息（不包含请求体）
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					// 处理敏感信息（如 Authorization 头）
					for idx, header := range headers {
						current := strings.Split(header, ":")
						if current[0] == "Authorization" {
							headers[idx] = current[0] + ": *"
						}
					}
					fields = append(fields, zap.Strings("headers", headers))
				}

				// 记录错误日志
				logging.Context(ctx).Error(fmt.Sprintf("[Recovery] %s panic recovered", time.Now().Format("2006/01/02 - 15:04:05")), fields...)
				// 返回 500 错误响应给客户端
				util.ResError(c, errors.InternalServerError("", "Internal server error, please try again later"))
			}
		}()

		// 继续处理后续的中间件和路由处理函数
		c.Next()
	}
}
