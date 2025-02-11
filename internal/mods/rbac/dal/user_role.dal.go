// Package dal 实现数据访问层（Data Access Layer）的功能
package dal

import (
	"context"
	"fmt"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetUserRoleDB 获取用户角色数据库实例
// ctx: 上下文
// defDB: 默认数据库连接
// 返回: 配置好的 GORM 数据库实例
func GetUserRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.UserRole))
}

// UserRole 用户角色数据访问结构体
type UserRole struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 查询用户角色列表
// ctx: 上下文
// params: 查询参数
// opts: 查询选项，可选参数
// 返回: 查询结果和可能的错误
func (a *UserRole) Query(ctx context.Context, params schema.UserRoleQueryParam, opts ...schema.UserRoleQueryOptions) (*schema.UserRoleQueryResult, error) {
	var opt schema.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 构建基础查询
	db := a.DB.Table(fmt.Sprintf("%s AS a", new(schema.UserRole).TableName()))

	// 如果需要关联角色表
	if opt.JoinRole {
		db = db.Joins(fmt.Sprintf("left join %s b on a.role_id=b.id", new(schema.Role).TableName()))
		db = db.Select("a.*,b.name as role_name")
	}

	// 添加查询条件
	if v := params.InUserIDs; len(v) > 0 {
		db = db.Where("a.user_id IN (?)", v)
	}
	if v := params.UserID; len(v) > 0 {
		db = db.Where("a.user_id = ?", v)
	}
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("a.role_id = ?", v)
	}

	// 执行分页查询
	var list schema.UserRoles
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.UserRoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 获取指定ID的用户角色
// ctx: 上下文
// id: 用户角色ID
// opts: 查询选项
// 返回: 用户角色信息和可能的错误
func (a *UserRole) Get(ctx context.Context, id string, opts ...schema.UserRoleQueryOptions) (*schema.UserRole, error) {
	var opt schema.UserRoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.UserRole)
	ok, err := util.FindOne(ctx, GetUserRoleDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查指定ID的用户角色是否存在
// ctx: 上下文
// id: 用户角色ID
// 返回: 是否存在和可能的错误
func (a *UserRole) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetUserRoleDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新的用户角色
// ctx: 上下文
// item: 要创建的用户角色信息
// 返回: 可能的错误
func (a *UserRole) Create(ctx context.Context, item *schema.UserRole) error {
	result := GetUserRoleDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新用户角色信息
// ctx: 上下文
// item: 要更新的用户角色信息
// 返回: 可能的错误
func (a *UserRole) Update(ctx context.Context, item *schema.UserRole) error {
	result := GetUserRoleDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定ID的用户角色
// ctx: 上下文
// id: 要删除的用户角色ID
// 返回: 可能的错误
func (a *UserRole) Delete(ctx context.Context, id string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}

// DeleteByUserID 删除指定用户ID关联的所有角色
// ctx: 上下文
// userID: 用户ID
// 返回: 可能的错误
func (a *UserRole) DeleteByUserID(ctx context.Context, userID string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("user_id=?", userID).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}

// DeleteByRoleID 删除指定角色ID关联的所有用户角色记录
// ctx: 上下文
// roleID: 角色ID
// 返回: 可能的错误
func (a *UserRole) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetUserRoleDB(ctx, a.DB).Where("role_id=?", roleID).Delete(new(schema.UserRole))
	return errors.WithStack(result.Error)
}
