// Package rbac 实现了基于角色的访问控制系统
package rbac

import (
	"context"
	"path/filepath"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/api"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RBAC 结构体是整个RBAC模块的核心，包含了所有必要的组件
type RBAC struct {
	DB        *gorm.DB    // 数据库连接实例
	MenuAPI   *api.Menu   // 菜单相关的API处理器
	RoleAPI   *api.Role   // 角色相关的API处理器
	UserAPI   *api.User   // 用户相关的API处理器
	LoginAPI  *api.Login  // 登录相关的API处理器
	LoggerAPI *api.Logger // 日志相关的API处理器
	Casbinx   *Casbinx    // Casbin权限管理实例
}

// AutoMigrate 自动迁移数据库表结构
// 根据定义的schema自动创建或更新数据库表
func (a *RBAC) AutoMigrate(ctx context.Context) error {
	return a.DB.AutoMigrate(
		new(schema.Menu),         // 菜单表
		new(schema.MenuResource), // 菜单资源表
		new(schema.Role),         // 角色表
		new(schema.RoleMenu),     // 角色菜单关联表
		new(schema.User),         // 用户表
		new(schema.UserRole),     // 用户角色关联表
	)
}

// Init 初始化RBAC模块
// 包括数据库迁移、加载Casbin策略和初始化菜单数据
func (a *RBAC) Init(ctx context.Context) error {
	// 如果配置了自动迁移，则执行数据库迁移
	if config.C.Storage.DB.AutoMigrate {
		if err := a.AutoMigrate(ctx); err != nil {
			return err
		}
	}

	// 加载Casbin权限策略
	if err := a.Casbinx.Load(ctx); err != nil {
		return err
	}

	// 如果配置了菜单文件，则从文件初始化菜单数据
	if name := config.C.General.MenuFile; name != "" {
		fullPath := filepath.Join(config.C.General.WorkDir, name)
		if err := a.MenuAPI.MenuBIZ.InitFromFile(ctx, fullPath); err != nil {
			logging.Context(ctx).Error("初始化菜单数据失败", zap.Error(err), zap.String("file", fullPath))
		}
	}

	return nil
}

// RegisterV1Routers 注册V1版本的API路由
// 设置所有的HTTP接口路由
func (a *RBAC) RegisterV1Routers(ctx context.Context, v1 *gin.RouterGroup) error {
	// 验证码相关路由
	captcha := v1.Group("captcha")
	{
		captcha.GET("id", a.LoginAPI.GetCaptcha)         // 获取验证码ID
		captcha.GET("image", a.LoginAPI.ResponseCaptcha) // 获取验证码图片
	}

	// 登录路由
	v1.POST("login", a.LoginAPI.Login)

	// 当前用户相关路由
	current := v1.Group("current")
	{
		current.POST("refresh-token", a.LoginAPI.RefreshToken) // 刷新令牌
		current.GET("user", a.LoginAPI.GetUserInfo)            // 获取用户信息
		current.GET("menus", a.LoginAPI.QueryMenus)            // 查询用户菜单
		current.PUT("password", a.LoginAPI.UpdatePassword)     // 更新密码
		current.PUT("user", a.LoginAPI.UpdateUser)             // 更新用户信息
		current.POST("logout", a.LoginAPI.Logout)              // 退出登录
	}

	// 菜单管理路由
	menu := v1.Group("menus")
	{
		menu.GET("", a.MenuAPI.Query)        // 查询菜单列表
		menu.GET(":id", a.MenuAPI.Get)       // 获取指定菜单
		menu.POST("", a.MenuAPI.Create)      // 创建菜单
		menu.PUT(":id", a.MenuAPI.Update)    // 更新菜单
		menu.DELETE(":id", a.MenuAPI.Delete) // 删除菜单
	}

	// 角色管理路由
	role := v1.Group("roles")
	{
		role.GET("", a.RoleAPI.Query)        // 查询角色列表
		role.GET(":id", a.RoleAPI.Get)       // 获取指定角色
		role.POST("", a.RoleAPI.Create)      // 创建角色
		role.PUT(":id", a.RoleAPI.Update)    // 更新角色
		role.DELETE(":id", a.RoleAPI.Delete) // 删除角色
	}

	// 用户管理路由
	user := v1.Group("users")
	{
		user.GET("", a.UserAPI.Query)                        // 查询用户列表
		user.GET(":id", a.UserAPI.Get)                       // 获取指定用户
		user.POST("", a.UserAPI.Create)                      // 创建用户
		user.PUT(":id", a.UserAPI.Update)                    // 更新用户
		user.DELETE(":id", a.UserAPI.Delete)                 // 删除用户
		user.PATCH(":id/reset-pwd", a.UserAPI.ResetPassword) // 重置用户密码
	}

	// 日志查询路由
	logger := v1.Group("loggers")
	{
		logger.GET("", a.LoggerAPI.Query) // 查询日志列表
	}

	return nil
}

// Release 释放RBAC模块资源
// 主要用于清理和释放Casbin相关资源
func (a *RBAC) Release(ctx context.Context) error {
	if err := a.Casbinx.Release(ctx); err != nil {
		return err
	}
	return nil
}
