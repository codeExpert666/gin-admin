// Package dal 实现数据访问层（Data Access Layer）的功能
package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetRoleMenuDB 获取角色菜单的数据库实例
// ctx: 上下文信息
// defDB: 默认的数据库连接
// 返回: 配置好的 GORM 数据库实例
func GetRoleMenuDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.RoleMenu))
}

// RoleMenu 结构体定义了角色菜单的数据访问对象
// 用于处理 RBAC（基于角色的访问控制）中角色与菜单的关联关系
type RoleMenu struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 查询角色菜单列表
// ctx: 上下文信息
// params: 查询参数
// opts: 可选的查询选项
// 返回: 查询结果和可能的错误
func (a *RoleMenu) Query(ctx context.Context, params schema.RoleMenuQueryParam, opts ...schema.RoleMenuQueryOptions) (*schema.RoleMenuQueryResult, error) {
	var opt schema.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 获取数据库实例并构建查询条件
	db := GetRoleMenuDB(ctx, a.DB)
	if v := params.RoleID; len(v) > 0 {
		db = db.Where("role_id = ?", v)
	}

	var list schema.RoleMenus
	// 执行分页查询
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.RoleMenuQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 根据ID获取指定的角色菜单记录
// ctx: 上下文信息
// id: 记录ID
// opts: 可选的查询选项
// 返回: 角色菜单记录和可能的错误
func (a *RoleMenu) Get(ctx context.Context, id string, opts ...schema.RoleMenuQueryOptions) (*schema.RoleMenu, error) {
	var opt schema.RoleMenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.RoleMenu)
	// 查找指定ID的记录
	ok, err := util.FindOne(ctx, GetRoleMenuDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查指定ID的角色菜单记录是否存在
// ctx: 上下文信息
// id: 记录ID
// 返回: 是否存在的布尔值和可能的错误
func (a *RoleMenu) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleMenuDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Create 创建新的角色菜单记录
// ctx: 上下文信息
// item: 要创建的角色菜单数据
// 返回: 可能的错误
func (a *RoleMenu) Create(ctx context.Context, item *schema.RoleMenu) error {
	result := GetRoleMenuDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新指定的角色菜单记录
// ctx: 上下文信息
// item: 要更新的角色菜单数据
// 返回: 可能的错误
func (a *RoleMenu) Update(ctx context.Context, item *schema.RoleMenu) error {
	// 更新除 created_at 外的所有字段
	result := GetRoleMenuDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定ID的角色菜单记录
// ctx: 上下文信息
// id: 要删除的记录ID
// 返回: 可能的错误
func (a *RoleMenu) Delete(ctx context.Context, id string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByRoleID 根据角色ID删除相关的所有角色菜单记录
// ctx: 上下文信息
// roleID: 角色ID
// 返回: 可能的错误
func (a *RoleMenu) DeleteByRoleID(ctx context.Context, roleID string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("role_id=?", roleID).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID 根据菜单ID删除相关的所有角色菜单记录
// ctx: 上下文信息
// menuID: 菜单ID
// 返回: 可能的错误
func (a *RoleMenu) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetRoleMenuDB(ctx, a.DB).Where("menu_id=?", menuID).Delete(new(schema.RoleMenu))
	return errors.WithStack(result.Error)
}
