package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Role 角色管理结构体，实现RBAC（基于角色的访问控制）
type Role struct {
	Cache       cachex.Cacher // 缓存接口，用于处理缓存相关操作
	Trans       *util.Trans   // 事务管理器，用于处理数据库事务
	RoleDAL     *dal.Role     // 角色数据访问层
	RoleMenuDAL *dal.RoleMenu // 角色菜单关联数据访问层
	UserRoleDAL *dal.UserRole // 用户角色关联数据访问层
}

// Query 查询角色列表
// @param ctx 上下文
// @param params 查询参数
// @return 查询结果和可能的错误
func (a *Role) Query(ctx context.Context, params schema.RoleQueryParam) (*schema.RoleQueryResult, error) {
	// 默认启用分页
	params.Pagination = true

	// 如果是选择类型的查询，则只返回 id 和 name 字段，且不分页
	var selectFields []string
	if params.ResultType == schema.RoleResultTypeSelect {
		params.Pagination = false
		selectFields = []string{"id", "name"}
	}

	// 执行查询，设置排序规则：按序号降序，创建时间降序
	result, err := a.RoleDAL.Query(ctx, params, schema.RoleQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "sequence", Direction: util.DESC},
				{Field: "created_at", Direction: util.DESC},
			},
			SelectFields: selectFields,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get 获取指定ID的角色详情
// @param ctx 上下文
// @param id 角色ID
// @return 角色信息和可能的错误
func (a *Role) Get(ctx context.Context, id string) (*schema.Role, error) {
	// 查询角色基本信息
	role, err := a.RoleDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if role == nil {
		return nil, errors.NotFound("", "Role not found")
	}

	// 查询角色关联的菜单信息
	roleMenuResult, err := a.RoleMenuDAL.Query(ctx, schema.RoleMenuQueryParam{
		RoleID: id,
	})
	if err != nil {
		return nil, err
	}
	role.Menus = roleMenuResult.Data

	return role, nil
}

// Create 创建新角色
// @param ctx 上下文
// @param formItem 角色表单数据
// @return 创建的角色信息和可能的错误
func (a *Role) Create(ctx context.Context, formItem *schema.RoleForm) (*schema.Role, error) {
	// 检查角色代码是否已存在
	if exists, err := a.RoleDAL.ExistsCode(ctx, formItem.Code); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "Role code already exists")
	}

	// 创建角色基本信息
	role := &schema.Role{
		ID:        util.NewXID(), // 生成新的唯一ID
		CreatedAt: time.Now(),    // 设置创建时间
	}
	if err := formItem.FillTo(role); err != nil {
		return nil, err
	}

	// 使用事务创建角色及其关联的菜单信息
	err := a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 创建角色基本信息
		if err := a.RoleDAL.Create(ctx, role); err != nil {
			return err
		}

		// 创建角色关联的菜单信息
		for _, roleMenu := range formItem.Menus {
			roleMenu.ID = util.NewXID()
			roleMenu.RoleID = role.ID
			roleMenu.CreatedAt = time.Now()
			if err := a.RoleMenuDAL.Create(ctx, roleMenu); err != nil {
				return err
			}
		}
		// 同步到 Casbin 权限管理器
		return a.syncToCasbin(ctx)
	})
	if err != nil {
		return nil, err
	}
	role.Menus = formItem.Menus

	return role, nil
}

// Update 更新指定角色信息
// @param ctx 上下文
// @param id 角色ID
// @param formItem 角色更新表单
// @return 可能的错误
func (a *Role) Update(ctx context.Context, id string, formItem *schema.RoleForm) error {
	// 检查角色是否存在并验证角色代码
	role, err := a.RoleDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if role == nil {
		return errors.NotFound("", "Role not found")
	} else if role.Code != formItem.Code {
		// 如果修改了角色代码，需要检查新代码是否已存在
		if exists, err := a.RoleDAL.ExistsCode(ctx, formItem.Code); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "Role code already exists")
		}
	}

	// 更新角色信息
	if err := formItem.FillTo(role); err != nil {
		return err
	}
	role.UpdatedAt = time.Now()

	// 使用事务更新角色及其关联的菜单信息
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 更新角色基本信息
		if err := a.RoleDAL.Update(ctx, role); err != nil {
			return err
		}
		// 删除原有的角色菜单关联
		if err := a.RoleMenuDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		// 创建新的角色菜单关联
		for _, roleMenu := range formItem.Menus {
			if roleMenu.ID == "" {
				roleMenu.ID = util.NewXID()
			}
			roleMenu.RoleID = role.ID
			if roleMenu.CreatedAt.IsZero() {
				roleMenu.CreatedAt = time.Now()
			}
			roleMenu.UpdatedAt = time.Now()
			if err := a.RoleMenuDAL.Create(ctx, roleMenu); err != nil {
				return err
			}
		}
		// 同步到 Casbin 权限管理器
		return a.syncToCasbin(ctx)
	})
}

// Delete 删除指定角色
// @param ctx 上下文
// @param id 角色ID
// @return 可能的错误
func (a *Role) Delete(ctx context.Context, id string) error {
	// 检查角色是否存在
	exists, err := a.RoleDAL.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "Role not found")
	}

	// 使用事务删除角色及其关联数据
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 删除角色基本信息
		if err := a.RoleDAL.Delete(ctx, id); err != nil {
			return err
		}
		// 删除角色菜单关联
		if err := a.RoleMenuDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		// 删除用户角色关联
		if err := a.UserRoleDAL.DeleteByRoleID(ctx, id); err != nil {
			return err
		}

		// 同步到 Casbin 权限管理器
		return a.syncToCasbin(ctx)
	})
}

// syncToCasbin 同步角色数据到 Casbin 权限管理器
// @param ctx 上下文
// @return 可能的错误
func (a *Role) syncToCasbin(ctx context.Context) error {
	return a.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
