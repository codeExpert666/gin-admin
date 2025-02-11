// Package util 提供了一组用于处理 HTTP 请求和响应的工具函数
package util

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

// GetToken 从请求头或查询参数中获取访问令牌
// 首先尝试从 Authorization 头中获取 Bearer token
// 如果没有找到,则尝试从 URL 查询参数 accessToken 中获取
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = auth
	}

	if token == "" {
		token = c.Query("accessToken")
	}

	return token
}

// GetBodyData 从上下文中获取请求体数据
// 返回请求体的原始字节数据
// 如果数据不存在或类型不匹配则返回 nil
func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

// ParseJSON 将请求体中的 JSON 数据解析到指定的结构体中
// 参数:
//   - c: Gin 上下文
//   - obj: 目标结构体指针
//
// 返回:
//   - error: 解析失败时返回错误信息
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.BadRequest("", "Failed to parse json: %s", err.Error())
	}
	return nil
}

// ParseQuery 将 URL 查询参数解析到指定的结构体中
// 参数:
//   - c: Gin 上下文
//   - obj: 目标结构体指针
//
// 返回:
//   - error: 解析失败时返回错误信息
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.BadRequest("", "Failed to parse query: %s", err.Error())
	}
	return nil
}

// ParseForm 将表单数据解析到指定的结构体中
// 支持 multipart/form-data 和 application/x-www-form-urlencoded 格式
// 参数:
//   - c: Gin 上下文
//   - obj: 目标结构体指针
//
// 返回:
//   - error: 解析失败时返回错误信息
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.BadRequest("", "Failed to parse form: %s", err.Error())
	}
	return nil
}

// ResJSON 返回 JSON 格式的响应数据
// 参数:
//   - c: Gin 上下文
//   - status: HTTP 状态码
//   - v: 要序列化为 JSON 的数据
//
// 注意: 该函数会终止后续的中间件执行
func ResJSON(c *gin.Context, status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

// ResSuccess 返回成功的 JSON 响应
// 将数据包装在 ResponseResult 结构中,并设置 Success 为 true
// 参数:
//   - c: Gin 上下文
//   - v: 响应数据
func ResSuccess(c *gin.Context, v interface{}) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
	})
}

// ResOK 返回一个简单的成功响应
// 不包含具体数据,只返回 success: true
func ResOK(c *gin.Context) {
	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
	})
}

// ResPage 返回分页数据的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - v: 分页数据列表
//   - pr: 分页结果信息,包含总数等
//
// 注意: 如果数据为空,会初始化为空数组
func ResPage(c *gin.Context, v interface{}, pr *PaginationResult) {
	var total int64
	if pr != nil {
		total = pr.Total
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	if reflectValue.IsNil() {
		v = make([]interface{}, 0)
	}

	ResJSON(c, http.StatusOK, ResponseResult{
		Success: true,
		Data:    v,
		Total:   total,
	})
}

// ResError 返回错误响应
// 参数:
//   - c: Gin 上下文
//   - err: 错误信息
//   - status: 可选的 HTTP 状态码
//
// 功能:
//  1. 将普通错误转换为自定义错误类型
//  2. 对于 500 以上的错误码会记录详细日志
//  3. 统一错误响应格式
func ResError(c *gin.Context, err error, status ...int) {
	var ierr *errors.Error
	if e, ok := errors.As(err); ok {
		ierr = e
	} else {
		ierr = errors.FromError(errors.InternalServerError("", err.Error()))
	}

	code := int(ierr.Code)
	if len(status) > 0 {
		code = status[0]
	}

	if code >= 500 {
		ctx := c.Request.Context()
		ctx = logging.NewTag(ctx, logging.TagKeySystem)
		ctx = logging.NewStack(ctx, fmt.Sprintf("%+v", err))
		logging.Context(ctx).Error("Internal server error", zap.Error(err))
		ierr.Detail = http.StatusText(http.StatusInternalServerError)
	}

	ierr.Code = int32(code)
	ResJSON(c, code, ResponseResult{Error: ierr})
}
