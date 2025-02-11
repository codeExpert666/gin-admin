package biz

import (
	"context"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/crypto/hash"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// User 结构体实现了 RBAC（基于角色的访问控制）中用户管理的业务逻辑
type User struct {
	Cache       cachex.Cacher // 缓存接口，用于处理用户相关的缓存操作
	Trans       *util.Trans   // 事务管理器，用于处理需要事务的数据库操作
	UserDAL     *dal.User     // 用户数据访问层，处理用户基本信息的数据库操作
	UserRoleDAL *dal.UserRole // 用户角色关联的数据访问层，处理用户和角色关联的数据库操作
}

// Query 方法用于根据查询参数获取用户列表
// ctx: 上下文信息
// params: 查询参数，包含分页、筛选条件等
func (a *User) Query(ctx context.Context, params schema.UserQueryParam) (*schema.UserQueryResult, error) {
	// 设置分页
	params.Pagination = true

	// 查询用户列表，按创建时间降序排序，并排除密码字段
	result, err := a.UserDAL.Query(ctx, params, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC},
			},
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	}

	// 如果查询结果不为空，获取用户关联的角色信息
	if userIDs := result.Data.ToIDs(); len(userIDs) > 0 {
		userRoleResult, err := a.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
			InUserIDs: userIDs,
		}, schema.UserRoleQueryOptions{
			JoinRole: true,
		})
		if err != nil {
			return nil, err
		}
		// 将用户角色信息映射到对应的用户
		userRolesMap := userRoleResult.Data.ToUserIDMap()
		for _, user := range result.Data {
			user.Roles = userRolesMap[user.ID]
		}
	}

	return result, nil
}

// Get 方法用于获取指定ID的用户详细信息
// ctx: 上下文信息
// id: 用户ID
func (a *User) Get(ctx context.Context, id string) (*schema.User, error) {
	// 查询用户基本信息，排除密码字段
	user, err := a.UserDAL.Get(ctx, id, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	// 查询用户关联的角色信息
	userRoleResult, err := a.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

// Create 方法用于创建新用户
// ctx: 上下文信息
// formItem: 用户表单数据
func (a *User) Create(ctx context.Context, formItem *schema.UserForm) (*schema.User, error) {
	// 检查用户名是否已存在
	existsUsername, err := a.UserDAL.ExistsUsername(ctx, formItem.Username)
	if err != nil {
		return nil, err
	} else if existsUsername {
		return nil, errors.BadRequest("", "Username already exists")
	}

	// 创建新用户对象
	user := &schema.User{
		ID:        util.NewXID(), // 生成唯一ID
		CreatedAt: time.Now(),    // 设置创建时间
	}

	// 如果未设置密码，使用默认密码
	if formItem.Password == "" {
		formItem.Password = config.C.General.DefaultLoginPwd
	}

	// 将表单数据填充到用户对象
	if err := formItem.FillTo(user); err != nil {
		return nil, err
	}

	// 在事务中执行创建操作
	err = a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 创建用户基本信息
		if err := a.UserDAL.Create(ctx, user); err != nil {
			return err
		}

		// 创建用户角色关联
		for _, userRole := range formItem.Roles {
			userRole.ID = util.NewXID()
			userRole.UserID = user.ID
			userRole.CreatedAt = time.Now()
			if err := a.UserRoleDAL.Create(ctx, userRole); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	user.Roles = formItem.Roles

	return user, nil
}

// Update 方法用于更新指定用户的信息
// ctx: 上下文信息
// id: 用户ID
// formItem: 更新的表单数据
func (a *User) Update(ctx context.Context, id string, formItem *schema.UserForm) error {
	// 检查用户是否存在
	user, err := a.UserDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	} else if user.Username != formItem.Username {
		// 如果修改了用户名，检查新用户名是否已存在
		existsUsername, err := a.UserDAL.ExistsUsername(ctx, formItem.Username)
		if err != nil {
			return err
		} else if existsUsername {
			return errors.BadRequest("", "Username already exists")
		}
	}

	// 更新用户信息
	if err := formItem.FillTo(user); err != nil {
		return err
	}
	user.UpdatedAt = time.Now()

	// 在事务中执行更新操作
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 更新用户基本信息
		if err := a.UserDAL.Update(ctx, user); err != nil {
			return err
		}

		// 更新用户角色关联
		// 先删除原有的角色关联
		if err := a.UserRoleDAL.DeleteByUserID(ctx, id); err != nil {
			return err
		}
		// 创建新的角色关联
		for _, userRole := range formItem.Roles {
			if userRole.ID == "" {
				userRole.ID = util.NewXID()
			}
			userRole.UserID = user.ID
			if userRole.CreatedAt.IsZero() {
				userRole.CreatedAt = time.Now()
			}
			userRole.UpdatedAt = time.Now()
			if err := a.UserRoleDAL.Create(ctx, userRole); err != nil {
				return err
			}
		}

		// 删除用户缓存
		return a.Cache.Delete(ctx, config.CacheNSForUser, id)
	})
}

// Delete 方法用于删除指定用户
// ctx: 上下文信息
// id: 要删除的用户ID
func (a *User) Delete(ctx context.Context, id string) error {
	// 检查用户是否存在
	exists, err := a.UserDAL.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	// 在事务中执行删除操作
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 删除用户基本信息
		if err := a.UserDAL.Delete(ctx, id); err != nil {
			return err
		}
		// 删除用户角色关联
		if err := a.UserRoleDAL.DeleteByUserID(ctx, id); err != nil {
			return err
		}
		// 删除用户缓存
		return a.Cache.Delete(ctx, config.CacheNSForUser, id)
	})
}

// ResetPassword 方法用于重置用户密码为系统默认密码
// ctx: 上下文信息
// id: 用户ID
func (a *User) ResetPassword(ctx context.Context, id string) error {
	// 检查用户是否存在
	exists, err := a.UserDAL.Exists(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return errors.NotFound("", "User not found")
	}

	// 生成新的密码哈希
	hashPass, err := hash.GeneratePassword(config.C.General.DefaultLoginPwd)
	if err != nil {
		return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
	}

	// 在事务中执行密码更新
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.UserDAL.UpdatePasswordByID(ctx, id, hashPass); err != nil {
			return err
		}
		return nil
	})
}

// GetRoleIDs 方法用于获取指定用户的所有角色ID
// ctx: 上下文信息
// id: 用户ID
// 返回值: 角色ID切片和可能的错误
func (a *User) GetRoleIDs(ctx context.Context, id string) ([]string, error) {
	// 查询用户的角色关联信息，只选择角色ID字段
	userRoleResult, err := a.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: id,
	}, schema.UserRoleQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"role_id"},
		},
	})
	if err != nil {
		return nil, err
	}
	return userRoleResult.Data.ToRoleIDs(), nil
}
