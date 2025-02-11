package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// 定义日志标签常量，用于标识不同类型的日志
const (
	TagKeyMain     = "main"     // 主程序日志
	TagKeyRecovery = "recovery" // 恢复相关日志
	TagKeyRequest  = "request"  // 请求相关日志
	TagKeyLogin    = "login"    // 登录相关日志
	TagKeyLogout   = "logout"   // 登出相关日志
	TagKeySystem   = "system"   // 系统相关日志
	TagKeyOperate  = "operate"  // 操作相关日志
)

// 定义用于 context 中存储各种日志相关信息的 key 类型
type (
	ctxLoggerKey  struct{} // 存储 logger 实例的 key
	ctxTraceIDKey struct{} // 存储追踪 ID 的 key
	ctxUserIDKey  struct{} // 存储用户 ID 的 key
	ctxTagKey     struct{} // 存储日志标签的 key
	ctxStackKey   struct{} // 存储堆栈信息的 key
)

// NewLogger 创建一个新的带有 logger 的 context
// 参数:
// - ctx: 原始 context
// - logger: zap logger 实例
// 返回: 新的包含 logger 的 context
func NewLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// FromLogger 从 context 中获取 logger
// 如果 context 中没有 logger，则返回全局默认的 logger
func FromLogger(ctx context.Context) *zap.Logger {
	v := ctx.Value(ctxLoggerKey{})
	if v != nil {
		if vv, ok := v.(*zap.Logger); ok {
			return vv
		}
	}
	return zap.L()
}

// NewTraceID 在 context 中设置追踪 ID
func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxTraceIDKey{}, traceID)
}

// FromTraceID 从 context 中获取追踪 ID
// 如果不存在则返回空字符串
func FromTraceID(ctx context.Context) string {
	v := ctx.Value(ctxTraceIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewUserID 在 context 中设置用户 ID
func NewUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

// FromUserID 从 context 中获取用户 ID
// 如果不存在则返回空字符串
func FromUserID(ctx context.Context) string {
	v := ctx.Value(ctxUserIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewTag 在 context 中设置日志标签
func NewTag(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, ctxTagKey{}, tag)
}

// FromTag 从 context 中获取日志标签
// 如果不存在则返回空字符串
func FromTag(ctx context.Context) string {
	v := ctx.Value(ctxTagKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewStack 在 context 中设置堆栈信息
func NewStack(ctx context.Context, stack string) context.Context {
	return context.WithValue(ctx, ctxStackKey{}, stack)
}

// FromStack 从 context 中获取堆栈信息
// 如果不存在则返回空字符串
func FromStack(ctx context.Context) string {
	v := ctx.Value(ctxStackKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Context 从给定的 context 中提取所有日志相关字段
// 并创建一个包含这些字段的新 logger
// 会自动添加 trace_id, user_id, tag 和 stack 等字段（如果存在）
func Context(ctx context.Context) *zap.Logger {
	var fields []zap.Field
	if v := FromTraceID(ctx); v != "" {
		fields = append(fields, zap.String("trace_id", v))
	}
	if v := FromUserID(ctx); v != "" {
		fields = append(fields, zap.String("user_id", v))
	}
	if v := FromTag(ctx); v != "" {
		fields = append(fields, zap.String("tag", v))
	}
	if v := FromStack(ctx); v != "" {
		fields = append(fields, zap.String("stack", v))
	}
	return FromLogger(ctx).With(fields...)
}

// PrintLogger 实现了一个简单的打印日志的接口
// 主要用于兼容一些需要 Printf 方法的场景
type PrintLogger struct{}

// Printf 实现了打印日志的方法，内部使用 zap 来记录日志
func (a *PrintLogger) Printf(format string, args ...interface{}) {
	zap.L().Info(fmt.Sprintf(format, args...))
}
