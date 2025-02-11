// Package rbac 实现了基于角色的访问控制(Role-Based Access Control)功能
package rbac

import (
	// 导入所需的包
	// api 层负责处理 HTTP 请求和响应
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/api"
	// biz 层负责处理业务逻辑
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/biz"
	// dal 层负责数据访问(Data Access Layer)
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	// wire 包用于依赖注入
	"github.com/google/wire"
)

// Set 是一个 Wire 提供者集合，用于依赖注入
// wire.NewSet 创建一个新的提供者集合，包含所有需要注入的结构体
// wire.Struct 用于注册结构体及其依赖关系，"*" 表示注入所有字段
var Set = wire.NewSet(
	// 核心结构体
	wire.Struct(new(RBAC), "*"),    // RBAC 主结构体
	wire.Struct(new(Casbinx), "*"), // Casbin 权限管理相关结构体

	// 菜单管理相关结构体
	wire.Struct(new(dal.Menu), "*"),         // 菜单数据访问层
	wire.Struct(new(biz.Menu), "*"),         // 菜单业务逻辑层
	wire.Struct(new(api.Menu), "*"),         // 菜单 API 层
	wire.Struct(new(dal.MenuResource), "*"), // 菜单资源数据访问层

	// 角色管理相关结构体
	wire.Struct(new(dal.Role), "*"),     // 角色数据访问层
	wire.Struct(new(biz.Role), "*"),     // 角色业务逻辑层
	wire.Struct(new(api.Role), "*"),     // 角色 API 层
	wire.Struct(new(dal.RoleMenu), "*"), // 角色菜单关联数据访问层

	// 用户管理相关结构体
	wire.Struct(new(dal.User), "*"),     // 用户数据访问层
	wire.Struct(new(biz.User), "*"),     // 用户业务逻辑层
	wire.Struct(new(api.User), "*"),     // 用户 API 层
	wire.Struct(new(dal.UserRole), "*"), // 用户角色关联数据访问层

	// 登录和日志相关结构体
	wire.Struct(new(biz.Login), "*"),  // 登录业务逻辑层
	wire.Struct(new(api.Login), "*"),  // 登录 API 层
	wire.Struct(new(api.Logger), "*"), // 日志 API 层
	wire.Struct(new(biz.Logger), "*"), // 日志业务逻辑层
	wire.Struct(new(dal.Logger), "*"), // 日志数据访问层
)
