// Package middleware 提供了各种 Gin 框架的中间件
package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
)

// CopyBodyConfig 定义了复制请求体中间件的配置项
type CopyBodyConfig struct {
	AllowedPathPrefixes []string // 允许复制请求体的路径前缀列表
	SkippedPathPrefixes []string // 跳过复制请求体的路径前缀列表
	MaxContentLen       int64    // 允许的请求体最大长度
}

// DefaultCopyBodyConfig 定义了默认的配置
var DefaultCopyBodyConfig = CopyBodyConfig{
	MaxContentLen: 32 << 20, // 设置默认最大请求体大小为 32MB
}

// CopyBody 使用默认配置创建一个复制请求体的中间件
func CopyBody() gin.HandlerFunc {
	return CopyBodyWithConfig(DefaultCopyBodyConfig)
}

// CopyBodyWithConfig 根据自定义配置创建一个复制请求体的中间件
// 这个中间件的主要作用是复制请求体内容，使其可以被多次读取
func CopyBodyWithConfig(config CopyBodyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求路径是否满足处理条件
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			c.Request.Body == nil {
			c.Next()
			return
		}

		var (
			requestBody []byte
			err         error
		)

		// 处理 gzip 压缩的请求体
		isGzip := false
		// 创建一个带有最大长度限制的请求体读取器
		safe := http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxContentLen)

		// 检查请求头是否包含 gzip 编码
		if c.GetHeader("Content-Encoding") == "gzip" {
			if reader, ierr := gzip.NewReader(safe); ierr == nil {
				isGzip = true
				requestBody, err = io.ReadAll(reader)
			}
		}

		// 如果不是 gzip 压缩，直接读取请求体
		if !isGzip {
			requestBody, err = io.ReadAll(safe)
		}

		// 处理请求体过大的错误
		if err != nil {
			util.ResError(c, errors.RequestEntityTooLarge("", "Request body too large, limit %d byte", config.MaxContentLen))
			return
		}

		// 关闭原始请求体
		c.Request.Body.Close()
		// 创建一个新的缓冲区，包含复制的请求体内容
		bf := bytes.NewBuffer(requestBody)
		// 替换原始请求体为新的可重复读取的请求体
		c.Request.Body = io.NopCloser(bf)
		// 将请求体内容保存到上下文中，供后续中间件或处理函数使用
		c.Set(util.ReqBodyKey, requestBody)
		// 调用下一个中间件或处理函数
		c.Next()
	}
}
