// Package errors 提供了一个返回详细请求错误信息的包
// 错误信息通常会被编码为 JSON 格式
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

// 定义一些来自 github.com/pkg/errors 包的别名函数
// 这样可以直接使用这些函数而不需要导入原始包
var (
	WithStack = errors.WithStack // 为错误添加堆栈信息
	Wrap      = errors.Wrap      // 包装错误并添加新的信息
	Wrapf     = errors.Wrapf     // 包装错误并添加格式化的新信息
	Is        = errors.Is        // 判断两个错误是否相等
	Errorf    = errors.Errorf    // 创建一个格式化的错误
)

// 定义常用的错误 ID 常量
const (
	DefaultBadRequestID            = "bad_request"              // 400 错误的默认 ID
	DefaultUnauthorizedID          = "unauthorized"             // 401 错误的默认 ID
	DefaultForbiddenID             = "forbidden"                // 403 错误的默认 ID
	DefaultNotFoundID              = "not_found"                // 404 错误的默认 ID
	DefaultMethodNotAllowedID      = "method_not_allowed"       // 405 错误的默认 ID
	DefaultTooManyRequestsID       = "too_many_requests"        // 429 错误的默认 ID
	DefaultRequestEntityTooLargeID = "request_entity_too_large" // 413 错误的默认 ID
	DefaultInternalServerErrorID   = "internal_server_error"    // 500 错误的默认 ID
	DefaultConflictID              = "conflict"                 // 409 错误的默认 ID
	DefaultRequestTimeoutID        = "request_timeout"          // 408 错误的默认 ID
)

// Error 定义了自定义错误结构体
// 实现了 error 接口，可以被 JSON 序列化
type Error struct {
	ID     string `json:"id,omitempty"`     // 错误标识符
	Code   int32  `json:"code,omitempty"`   // HTTP 状态码
	Detail string `json:"detail,omitempty"` // 错误详细信息
	Status string `json:"status,omitempty"` // HTTP 状态描述
}

// Error 实现 error 接口
// 返回 JSON 格式的错误字符串
func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// New 创建一个新的自定义错误
// id: 错误标识符
// detail: 错误详细信息
// code: HTTP 状态码
func New(id, detail string, code int32) error {
	return &Error{
		ID:     id,
		Code:   code,
		Detail: detail,
		Status: http.StatusText(int(code)),
	}
}

// Parse 尝试将 JSON 字符串解析为错误对象
// 如果解析失败，会将输入字符串设置为错误详情
func Parse(err string) *Error {
	e := new(Error)
	errr := json.Unmarshal([]byte(err), e)
	if errr != nil {
		e.Detail = err
	}
	return e
}

// BadRequest 生成 400 错误
// id: 错误标识符，如果为空则使用默认值
// format: 错误信息格式
// a: 格式化参数
func BadRequest(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultBadRequestID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusBadRequest,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusBadRequest),
	}
}

// Unauthorized 生成 401 未授权错误
// 用于表示请求需要用户认证
func Unauthorized(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultUnauthorizedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusUnauthorized,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden 生成 403 禁止访问错误
// 用于表示用户没有访问权限
func Forbidden(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultForbiddenID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusForbidden,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusForbidden),
	}
}

// NotFound 生成 404 未找到错误
// 用于表示请求的资源不存在
func NotFound(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultNotFoundID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusNotFound,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusNotFound),
	}
}

// MethodNotAllowed 生成 405 方法不允许错误
// 用于表示请求方法不被允许
func MethodNotAllowed(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultMethodNotAllowedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusMethodNotAllowed,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusMethodNotAllowed),
	}
}

// TooManyRequests 生成 429 请求过多错误
// 用于限流场景，表示客户端在给定时间发送了太多请求
func TooManyRequests(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultTooManyRequestsID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusTooManyRequests,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusTooManyRequests),
	}
}

// Timeout 生成 408 请求超时错误
// 用于表示服务器等待客户端发送请求时发生超时
func Timeout(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestTimeoutID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestTimeout,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestTimeout),
	}
}

// Conflict 生成 409 冲突错误
// 用于表示请求与服务器当前状态存在冲突
func Conflict(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultConflictID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusConflict,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusConflict),
	}
}

// RequestEntityTooLarge 生成 413 请求实体过大错误
// 用于表示请求的实体超过服务器允许的大小
func RequestEntityTooLarge(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestEntityTooLargeID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestEntityTooLarge,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestEntityTooLarge),
	}
}

// InternalServerError 生成 500 服务器内部错误
// 用于表示服务器遇到了意外的情况
func InternalServerError(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultInternalServerErrorID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

// Equal 比较两个错误是否相等
// 如果两个错误都是 *Error 类型，则比较它们的 Code
// 否则直接比较错误实例
func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return err1 == err2
	}

	if verr1.Code != verr2.Code {
		return false
	}

	return true
}

// FromError 尝试将 Go 的 error 转换为 *Error
// 如果输入的 error 已经是 *Error 类型，直接返回
// 否则将错误消息解析为新的 *Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*Error); ok && verr != nil {
		return verr
	}

	return Parse(err.Error())
}

// As 在错误链中查找第一个匹配 *Error 类型的错误
// 使用 errors.As 进行类型断言
func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var merr *Error
	if errors.As(err, &merr) {
		return merr, true
	}
	return nil, false
}

// MultiError 定义了一个可以存储多个错误的结构体
// 适用于需要收集多个错误的场景
type MultiError struct {
	lock   *sync.Mutex // 用于并发安全的互斥锁
	Errors []error     // 存储多个错误的切片
}

// NewMultiError 创建一个新的 MultiError 实例
func NewMultiError() *MultiError {
	return &MultiError{
		lock:   &sync.Mutex{},
		Errors: make([]error, 0),
	}
}

// Append 添加一个新的错误到错误列表中
// 注意：这个方法不是并发安全的
func (e *MultiError) Append(err error) {
	e.Errors = append(e.Errors, err)
}

// AppendWithLock 以并发安全的方式添加错误
// 使用互斥锁保证并发安全
func (e *MultiError) AppendWithLock(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Append(err)
}

// HasErrors 检查是否包含任何错误
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Error 实现 error 接口
// 返回 JSON 格式的错误信息
func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
