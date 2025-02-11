package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetMenuDB 获取菜单数据库实例
// ctx: 上下文
// defDB: 默认数据库连接
// 返回: 配置好的 GORM 数据库实例
func GetMenuDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Menu))
}

// Menu 菜单数据访问层结构体
// 实现了 RBAC（基于角色的访问控制）中菜单管理相关的数据库操作
type Menu struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 根据查询参数查询菜单列表
// ctx: 上下文
// params: 查询参数，包含过滤条件
// opts: 可选的查询选项
// 返回: 查询结果和可能的错误
func (a *Menu) Query(ctx context.Context, params schema.MenuQueryParam, opts ...schema.MenuQueryOptions) (*schema.MenuQueryResult, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	db := GetMenuDB(ctx, a.DB)

	// 根据传入的 ID 列表筛选
	if v := params.InIDs; len(v) > 0 {
		db = db.Where("id IN ?", v)
	}
	// 根据名称模糊查询
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	// 根据状态筛选
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	// 根据父级 ID 筛选
	if v := params.ParentID; len(v) > 0 {
		db = db.Where("parent_id = ?", v)
	}
	// 根据父级路径前缀筛选
	if v := params.ParentPathPrefix; len(v) > 0 {
		db = db.Where("parent_path LIKE ?", v+"%")
	}
	// 根据用户 ID 筛选可访问的菜单
	if v := params.UserID; len(v) > 0 {
		userRoleQuery := GetUserRoleDB(ctx, a.DB).Where("user_id = ?", v).Select("role_id")
		roleMenuQuery := GetRoleMenuDB(ctx, a.DB).Where("role_id IN (?)", userRoleQuery).Select("menu_id")
		db = db.Where("id IN (?)", roleMenuQuery)
	}
	// 根据角色 ID 筛选可访问的菜单
	if v := params.RoleID; len(v) > 0 {
		roleMenuQuery := GetRoleMenuDB(ctx, a.DB).Where("role_id = ?", v).Select("menu_id")
		db = db.Where("id IN (?)", roleMenuQuery)
	}

	var list schema.Menus
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.MenuQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 根据 ID 获取指定菜单
// ctx: 上下文
// id: 菜单 ID
// opts: 可选的查询选项
// 返回: 菜单信息和可能的错误
func (a *Menu) Get(ctx context.Context, id string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// GetByCodeAndParentID 根据编码和父级 ID 获取菜单
// ctx: 上下文
// code: 菜单编码
// parentID: 父级菜单 ID
// opts: 可选的查询选项
// 返回: 菜单信息和可能的错误
func (a *Menu) GetByCodeAndParentID(ctx context.Context, code, parentID string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, a.DB).Where("code=? AND parent_id=?", code, parentID), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// GetByNameAndParentID get the specified menu from the database.
func (a *Menu) GetByNameAndParentID(ctx context.Context, name, parentID string, opts ...schema.MenuQueryOptions) (*schema.Menu, error) {
	var opt schema.MenuQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.Menu)
	ok, err := util.FindOne(ctx, GetMenuDB(ctx, a.DB).Where("name=? AND parent_id=?", name, parentID), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Checks if the specified menu exists in the database.
func (a *Menu) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// Checks if a menu with the specified `code` exists under the specified `parentID` in the database.
func (a *Menu) ExistsCodeByParentID(ctx context.Context, code, parentID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, a.DB).Where("code=? AND parent_id=?", code, parentID))
	return ok, errors.WithStack(err)
}

// Checks if a menu with the specified `name` exists under the specified `parentID` in the database.
func (a *Menu) ExistsNameByParentID(ctx context.Context, name, parentID string) (bool, error) {
	ok, err := util.Exists(ctx, GetMenuDB(ctx, a.DB).Where("name=? AND parent_id=?", name, parentID))
	return ok, errors.WithStack(err)
}

// Create 创建新的菜单
// ctx: 上下文
// item: 要创建的菜单信息
// 返回: 可能的错误
func (a *Menu) Create(ctx context.Context, item *schema.Menu) error {
	result := GetMenuDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新指定菜单
// ctx: 上下文
// item: 要更新的菜单信息
// 返回: 可能的错误
func (a *Menu) Update(ctx context.Context, item *schema.Menu) error {
	result := GetMenuDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定菜单
// ctx: 上下文
// id: 要删除的菜单 ID
// 返回: 可能的错误
func (a *Menu) Delete(ctx context.Context, id string) error {
	result := GetMenuDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Menu))
	return errors.WithStack(result.Error)
}

// UpdateParentPath 更新菜单的父级路径
// ctx: 上下文
// id: 菜单 ID
// parentPath: 新的父级路径
// 返回: 可能的错误
func (a *Menu) UpdateParentPath(ctx context.Context, id, parentPath string) error {
	result := GetMenuDB(ctx, a.DB).Where("id=?", id).Update("parent_path", parentPath)
	return errors.WithStack(result.Error)
}

// UpdateStatusByParentPath 根据父级路径更新所有子菜单的状态
// ctx: 上下文
// parentPath: 父级路径
// status: 新的状态值
// 返回: 可能的错误
func (a *Menu) UpdateStatusByParentPath(ctx context.Context, parentPath, status string) error {
	result := GetMenuDB(ctx, a.DB).Where("parent_path like ?", parentPath+"%").Update("status", status)
	return errors.WithStack(result.Error)
}
