// Package promx 提供了与 Prometheus 监控系统集成的功能
package promx

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdapterGin 结构体是 Gin 框架的 Prometheus 适配器
// 它包装了 PrometheusWrapper，用于收集和导出监控指标
type AdapterGin struct {
	prom *PrometheusWrapper
}

// NewAdapterGin 创建一个新的 Gin 适配器实例
// 参数 p 是 PrometheusWrapper 的指针，用于处理实际的指标收集
func NewAdapterGin(p *PrometheusWrapper) *AdapterGin {
	return &AdapterGin{prom: p}
}

// Middleware 返回一个 Gin 中间件函数，用于收集 HTTP 请求的性能指标
// 参数:
//   - enable: 是否启用监控
//   - reqKey: 用于从 Gin 上下文中获取请求体大小的键名
func (a *AdapterGin) Middleware(enable bool, reqKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 如果监控未启用，直接调用下一个处理器
		if !enable {
			ctx.Next()
			return
		}

		// 记录请求开始时间
		start := time.Now()

		// 获取请求体大小（如果存在）
		recvBytes := 0
		if v, ok := ctx.Get(reqKey); ok {
			if b, ok := v.([]byte); ok {
				recvBytes = len(b)
			}
		}

		// 执行后续的处理器
		ctx.Next()

		// 计算请求处理耗时（毫秒）
		latency := float64(time.Since(start).Milliseconds())

		// 处理 URL 路径，将具体的参数值替换为参数名
		// 例如：/users/123 会被转换为 /users/:id
		p := ctx.Request.URL.Path
		for _, param := range ctx.Params {
			p = strings.Replace(p, param.Value, ":"+param.Key, -1)
		}

		// 记录请求的监控指标
		// 包括：路径、HTTP方法、状态码、响应大小、请求体大小和处理耗时
		a.prom.Log(p, ctx.Request.Method, fmt.Sprintf("%d", ctx.Writer.Status()),
			float64(ctx.Writer.Size()), float64(recvBytes), latency)
	}
}
