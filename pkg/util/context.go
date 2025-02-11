// package util 提供了一些通用的工具函数
package util

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"
	"gorm.io/gorm"
)

// 定义一组空结构体作为 context 的 key
// 使用空结构体作为 key 可以节省内存,因为空结构体不占用内存空间
type (
	traceIDCtx    struct{} // 用于存储请求追踪ID
	transCtx      struct{} // 用于存储数据库事务
	rowLockCtx    struct{} // 用于标记行锁
	userIDCtx     struct{} // 用于存储用户ID
	userTokenCtx  struct{} // 用于存储用户令牌
	isRootUserCtx struct{} // 用于标记是否为超级管理员
	userCacheCtx  struct{} // 用于存储用户缓存信息
)

// NewTraceID 创建一个新的上下文,并在其中存储请求追踪ID
// 参数:
//   - ctx: 父上下文
//   - traceID: 追踪ID字符串
//
// 返回: 包含追踪ID的新上下文
func NewTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDCtx{}, traceID)
}

// FromTraceID 从上下文中获取追踪ID
// 参数:
//   - ctx: 上下文对象
//
// 返回: 追踪ID字符串,如果不存在则返回空字符串
func FromTraceID(ctx context.Context) string {
	v := ctx.Value(traceIDCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// NewTrans 创建一个新的上下文,并在其中存储数据库事务对象
// 参数:
//   - ctx: 父上下文
//   - db: GORM数据库事务对象
//
// 返回: 包含数据库事务的新上下文
func NewTrans(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, transCtx{}, db)
}

// FromTrans 从上下文中获取数据库事务对象
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - *gorm.DB: 数据库事务对象
//   - bool: 是否成功获取到事务对象
func FromTrans(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(transCtx{})
	if v != nil {
		return v.(*gorm.DB), true
	}
	return nil, false
}

// NewRowLock 创建一个新的上下文,并在其中标记行锁状态
// 参数:
//   - ctx: 父上下文
//
// 返回: 包含行锁标记的新上下文
func NewRowLock(ctx context.Context) context.Context {
	return context.WithValue(ctx, rowLockCtx{}, true)
}

// FromRowLock 从上下文中获取行锁状态
// 参数:
//   - ctx: 上下文对象
//
// 返回: 是否启用了行锁
func FromRowLock(ctx context.Context) bool {
	v := ctx.Value(rowLockCtx{})
	return v != nil && v.(bool)
}

// NewUserID 创建一个新的上下文,并在其中存储用户ID
// 参数:
//   - ctx: 父上下文
//   - userID: 用户ID字符串
//
// 返回: 包含用户ID的新上下文
func NewUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtx{}, userID)
}

// FromUserID 从上下文中获取用户ID
// 参数:
//   - ctx: 上下文对象
//
// 返回: 用户ID字符串,如果不存在则返回空字符串
func FromUserID(ctx context.Context) string {
	v := ctx.Value(userIDCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// NewUserToken 创建一个新的上下文,并在其中存储用户令牌
// 参数:
//   - ctx: 父上下文
//   - userToken: 用户令牌字符串
//
// 返回: 包含用户令牌的新上下文
func NewUserToken(ctx context.Context, userToken string) context.Context {
	return context.WithValue(ctx, userTokenCtx{}, userToken)
}

// FromUserToken 从上下文中获取用户令牌
// 参数:
//   - ctx: 上下文对象
//
// 返回: 用户令牌字符串,如果不存在则返回空字符串
func FromUserToken(ctx context.Context) string {
	v := ctx.Value(userTokenCtx{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// NewIsRootUser 创建一个新的上下文,并在其中标记超级管理员状态
// 参数:
//   - ctx: 父上下文
//
// 返回: 包含超级管理员标记的新上下文
func NewIsRootUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, isRootUserCtx{}, true)
}

// FromIsRootUser 从上下文中获取是否为超级管理员
// 参数:
//   - ctx: 上下文对象
//
// 返回: 是否为超级管理员
func FromIsRootUser(ctx context.Context) bool {
	v := ctx.Value(isRootUserCtx{})
	return v != nil && v.(bool)
}

// UserCache 用户缓存对象,用于存储用户相关的缓存数据
type UserCache struct {
	RoleIDs []string `json:"rids"` // 用户角色ID列表
}

// ParseUserCache 将字符串解析为用户缓存对象
// 参数:
//   - s: JSON格式的字符串
//
// 返回: 解析后的UserCache对象
func ParseUserCache(s string) UserCache {
	var a UserCache
	if s == "" {
		return a
	}

	_ = json.Unmarshal([]byte(s), &a)
	return a
}

// String 将用户缓存对象转换为JSON字符串
// 返回: JSON格式的字符串
func (a UserCache) String() string {
	return json.MarshalToString(a)
}

// NewUserCache 创建一个新的上下文,并在其中存储用户缓存对象
// 参数:
//   - ctx: 父上下文
//   - userCache: 用户缓存对象
//
// 返回: 包含用户缓存的新上下文
func NewUserCache(ctx context.Context, userCache UserCache) context.Context {
	return context.WithValue(ctx, userCacheCtx{}, userCache)
}

// FromUserCache 从上下文中获取用户缓存对象
// 参数:
//   - ctx: 上下文对象
//
// 返回: 用户缓存对象,如果不存在则返回空对象
func FromUserCache(ctx context.Context) UserCache {
	v := ctx.Value(userCacheCtx{})
	if v != nil {
		return v.(UserCache)
	}
	return UserCache{}
}
