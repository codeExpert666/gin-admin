// Package middleware 提供了各种 Gin 框架的中间件功能
package middleware

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StaticConfig 定义了静态文件服务的配置结构
type StaticConfig struct {
	// SkippedPathPrefixes 定义了需要跳过的路径前缀列表
	// 比如设置为 ["/api"] 则所有以 /api 开头的请求都不会走静态文件处理
	SkippedPathPrefixes []string

	// Root 定义了静态文件的根目录路径
	// 所有的静态文件都将从这个目录下查找
	Root string
}

// StaticWithConfig 创建一个处理静态文件的中间件
// 该中间件会根据配置处理静态文件的请求
func StaticWithConfig(config StaticConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查当前请求路径是否在需要跳过的路径列表中
		// 如果是，则直接调用下一个中间件
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		// 获取请求的URL路径
		p := c.Request.URL.Path
		// 将URL路径转换为本地文件系统路径
		// filepath.Join 用于连接路径
		// filepath.FromSlash 用于将URL中的斜杠转换为系统对应的路径分隔符
		fpath := filepath.Join(config.Root, filepath.FromSlash(p))

		// 检查请求的文件是否存在
		_, err := os.Stat(fpath)
		// 如果文件不存在，则返回 index.html
		// 这是单页应用（SPA）的常见处理方式
		if err != nil && os.IsNotExist(err) {
			fpath = filepath.Join(config.Root, "index.html")
		}

		// 使用 Gin 的 File 方法发送文件
		c.File(fpath)
		// 终止后续中间件的执行
		c.Abort()
	}
}
