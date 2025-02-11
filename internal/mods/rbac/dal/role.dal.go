package dal

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetRoleDB 获取角色数据库实例
// ctx: 上下文信息
// defDB: 默认数据库连接
// 返回: 配置好的 GORM 数据库实例，指向 Role 模型
func GetRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Role))
}

// Role 角色数据访问层结构体
// 实现了对角色表的增删改查等基本操作
type Role struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 查询角色列表
// ctx: 上下文信息
// params: 查询参数，包含过滤条件、分页信息等
// opts: 可选的查询选项
// 返回: 查询结果（包含分页信息和角色列表）和可能的错误
func (a *Role) Query(ctx context.Context, params schema.RoleQueryParam, opts ...schema.RoleQueryOptions) (*schema.RoleQueryResult, error) {
	var opt schema.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 获取数据库实例并构建查询条件
	db := GetRoleDB(ctx, a.DB)
	// 根据ID列表过滤
	if v := params.InIDs; len(v) > 0 {
		db = db.Where("id IN (?)", v)
	}
	// 根据名称模糊查询
	if v := params.LikeName; len(v) > 0 {
		db = db.Where("name LIKE ?", "%"+v+"%")
	}
	// 根据状态过滤
	if v := params.Status; len(v) > 0 {
		db = db.Where("status = ?", v)
	}
	// 根据更新时间过滤
	if v := params.GtUpdatedAt; v != nil {
		db = db.Where("updated_at > ?", v)
	}

	var list schema.Roles
	// 执行分页查询
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	queryResult := &schema.RoleQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}

// Get 根据ID获取指定角色
// ctx: 上下文信息
// id: 角色ID
// opts: 可选的查询选项
// 返回: 角色信息和可能的错误
func (a *Role) Get(ctx context.Context, id string, opts ...schema.RoleQueryOptions) (*schema.Role, error) {
	var opt schema.RoleQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	item := new(schema.Role)
	// 查找指定ID的角色
	ok, err := util.FindOne(ctx, GetRoleDB(ctx, a.DB).Where("id=?", id), opt.QueryOptions, item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}
	return item, nil
}

// Exists 检查指定ID的角色是否存在
// ctx: 上下文信息
// id: 角色ID
// 返回: 是否存在及可能的错误
func (a *Role) Exists(ctx context.Context, id string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleDB(ctx, a.DB).Where("id=?", id))
	return ok, errors.WithStack(err)
}

// ExistsCode 检查指定编码的角色是否存在
// ctx: 上下文信息
// code: 角色编码
// 返回: 是否存在及可能的错误
func (a *Role) ExistsCode(ctx context.Context, code string) (bool, error) {
	ok, err := util.Exists(ctx, GetRoleDB(ctx, a.DB).Where("code=?", code))
	return ok, errors.WithStack(err)
}

// Create 创建新角色
// ctx: 上下文信息
// item: 要创建的角色信息
// 返回: 可能的错误
func (a *Role) Create(ctx context.Context, item *schema.Role) error {
	result := GetRoleDB(ctx, a.DB).Create(item)
	return errors.WithStack(result.Error)
}

// Update 更新角色信息
// ctx: 上下文信息
// item: 要更新的角色信息
// 返回: 可能的错误
func (a *Role) Update(ctx context.Context, item *schema.Role) error {
	// 更新除created_at外的所有字段
	result := GetRoleDB(ctx, a.DB).Where("id=?", item.ID).Select("*").Omit("created_at").Updates(item)
	return errors.WithStack(result.Error)
}

// Delete 删除指定角色
// ctx: 上下文信息
// id: 要删除的角色ID
// 返回: 可能的错误
func (a *Role) Delete(ctx context.Context, id string) error {
	result := GetRoleDB(ctx, a.DB).Where("id=?", id).Delete(new(schema.Role))
	return errors.WithStack(result.Error)
}
