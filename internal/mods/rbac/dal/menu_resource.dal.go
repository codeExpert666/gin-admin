// Package dal 实现数据访问层（Data Access Layer）的功能
package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetMenuResourceDB 获取菜单资源的数据库实例
// ctx: 上下文信息
// defDB: 默认的数据库连接
// 返回: 配置好的 GORM 数据库实例
func GetMenuResourceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.MenuResource))
}

// MenuResource 结构体用于实现菜单资源的数据库操作
type MenuResource struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 根据查询参数查询菜单资源列表
// ctx: 上下文信息
// params: 查询参数
// opts: 可选的查询选项
// 返回: 查询结果和可能的错误
func (a *MenuResource) Query(ctx context.Context, params schema.MenuResourceQueryParam, opts ...schema.MenuResourceQueryOptions) (*schema.MenuResourceQueryResult, error) {
	// 处理可选的查询选项
	var opt schema.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 获取数据库实例并构建查询条件
	db := GetMenuResourceDB(ctx, a.DB)
	if v := params.MenuID; len(v) > 0 {
		db = db.Where("menu_id = ?", v)
	}
	if v := params.MenuIDs; len(v) > 0 {
		db = db.Where("menu_id IN ?", v)
	}

	// 执行分页查询
	var list schema.MenuResources
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 封装查询结果
	queryResult := &schema.MenuResourceQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 根据ID获取指定的菜单资源
// ctx: 上下文信息
// id: 菜单资源ID
// opts: 可选的查询选项
// 返回: 菜单资源信息和可能的错误
func (a *MenuResource) Get(ctx context.Context, id string, opts ...schema.MenuResourceQueryOptions) (*schema.MenuResource, error) {
	var opt schema.MenuResourceQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.MenuResource)
	ok, err := util.FindOne(ctx, GetMenuResourceDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查指定ID的菜单资源是否存在
// ctx: 上下文信息
// id: 菜单资源ID
// 返回: 是否存在的布尔值和可能的错误
func (a *MenuResource) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuResourceDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsMethodPathByMenuID 检查指定菜单ID下是否存在特定方法和路径的资源
// ctx: 上下文信息
// method: HTTP方法
// path: 请求路径
// menuID: 菜单ID
// 返回: 是否存在的布尔值和可能的错误
func (a *MenuResource) ExistsMethodPathByMenuID(ctx context.Context, method, path, menuID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuResourceDB(ctx, a.DB).Where("method=? AND path=? AND menu_id=?", method, path, menuID))
	return ok, errors.WithStack(err)
}

// Create 创建新的菜单资源
// ctx: 上下文信息
// item: 要创建的菜单资源信息
// 返回: 可能的错误
func (a *MenuResource) Create(ctx context.Context, item *schema.MenuResource) error {
	result := GetMenuResourceDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新指定的菜单资源
// ctx: 上下文信息
// item: 要更新的菜单资源信息
// 返回: 可能的错误
func (a *MenuResource) Update(ctx context.Context, item *schema.MenuResource) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定ID的菜单资源
// ctx: 上下文信息
// id: 要删除的菜单资源ID
// 返回: 可能的错误
func (a *MenuResource) Delete(ctx context.Context, id string) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.MenuResource))
	return errors.WithStack(result.Error)
}

// DeleteByMenuID 根据菜单ID删除相关的菜单资源
// ctx: 上下文信息
// menuID: 菜单ID
// 返回: 可能的错误
func (a *MenuResource) DeleteByMenuID(ctx context.Context, menuID string) error {
	result := GetMenuResourceDB(ctx, a.DB).Where("menu_id=?", menuID).Delete(new(schema.MenuResource))
	return errors.WithStack(result.Error)
}
