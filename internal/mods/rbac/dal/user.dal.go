package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetUserDB 获取用户数据库实例
// ctx: 上下文
// defDB: 默认数据库连接
// 返回: 配置好的 GORM 数据库实例，已设置为操作 User 模型
func GetUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.User))
}

// User RBAC 系统中的用户管理结构体
// 封装了所有与用户相关的数据库操作
type User struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 根据查询参数查询用户列表
// ctx: 上下文
// params: 查询参数，包含分页、模糊搜索等条件
// opts: 可选的查询选项
// 返回: 查询结果（包含分页信息）和可能的错误
func (a *User) Query(ctx context.Context, params schema.UserQueryParam, opts ...schema.UserQueryOptions) (*schema.UserQueryResult, error) {
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 获取数据库实例并构建查询条件
	db := GetUserDB(ctx, a.DB)
	if v := params.LikeUsername; len(v) > 0 {
		db = db.Where("username LIKE ?", "%"+v+"%") // 用户名模糊查询
	}
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%") // 姓名模糊查询
	}
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v) // 状态精确查询
	}

	var list schema.Users
	// 执行分页查询
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.UserQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 根据用户ID获取用户信息
// ctx: 上下文
// id: 用户ID
// opts: 可选的查询选项
// 返回: 用户信息和可能的错误
func (a *User) Get(ctx context.Context, id string, opts ...schema.UserQueryOptions) (*schema.User, error) {
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.User)
	// 查找指定ID的用户
	ok, err := util.FindOne(ctx, GetUserDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil // 用户不存在返回 nil
	}
	return item, nil
}

// GetByUsername 根据用户名获取用户信息
// ctx: 上下文
// username: 用户名
// opts: 可选的查询选项
// 返回: 用户信息和可能的错误
func (a *User) GetByUsername(ctx context.Context, username string, opts ...schema.UserQueryOptions) (*schema.User, error) {
	var opt schema.UserQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.User)
	// 查找指定用户名的用户
	ok, err := util.FindOne(ctx, GetUserDB(ctx, a.DB).Where("username=?", username), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil // 用户不存在返回 nil
	}
	return item, nil
}

// Exists 检查指定ID的用户是否存在
// ctx: 上下文
// id: 用户ID
// 返回: 是否存在的布尔值和可能的错误
func (a *User) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsUsername 检查指定用户名的用户是否存在
// ctx: 上下文
// username: 用户名
// 返回: 是否存在的布尔值和可能的错误
func (a *User) ExistsUsername(ctx context.Context, username string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserDB(ctx, a.DB).Where("username=?", username))
	return ok, errors.WithStack(err)
}

// Create 创建新用户
// ctx: 上下文
// item: 要创建的用户信息
// 返回: 可能的错误
func (a *User) Create(ctx context.Context, item *schema.User) error {
	result := GetUserDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新用户信息
// ctx: 上下文
// item: 要更新的用户信息
// selectFields: 可选参数，指定要更新的字段，为空则更新所有字段（除 created_at 外）
// 返回: 可能的错误
func (a *User) Update(ctx context.Context, item *schema.User, selectFields ...string) error {
	db := GetUserDB(ctx, a.DB).Where("id=?", item.ID)
	if len(selectFields) > 0 {
		db = db.Select(selectFields) // 更新指定字段
	} else {
		db = db.Select("*").Omit("created_at") // 更新所有字段，除了 created_at
	}
	result := db.Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定用户
// ctx: 上下文
// id: 要删除的用户ID
// 返回: 可能的错误
func (a *User) Delete(ctx context.Context, id string) error {
	result := GetUserDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.User))
	return errors.WithStack(result.Error)
}

// UpdatePasswordByID 更新指定用户的密码
// ctx: 上下文
// id: 用户ID
// password: 新密码
// 返回: 可能的错误
func (a *User) UpdatePasswordByID(ctx context.Context, id string, password string) error {
	result := GetUserDB(ctx, a.DB).Where("id=?", id).Select("password").Updates(schema.User{Password: password})
	return errors.WithStack(result.Error)
}
