// Package mods 是模块包，负责管理和组织应用程序的不同模块
package mods

import (
	// 导入必要的包
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac" // 导入RBAC（基于角色的访问控制）模块
	"github.com/gin-gonic/gin"                              // Web框架
	"github.com/google/wire"                                // 依赖注入工具
)

// API路由前缀常量
const (
	apiPrefix = "/api/"
)

// Set 是wire provider的集合
// wire.NewSet用于依赖注入，将所有需要的依赖组织在一起
var Set = wire.NewSet(
	wire.Struct(new(Mods), "*"), // 注入Mods结构体
	rbac.Set,                    // 注入RBAC模块的依赖
)

// Mods 结构体定义了应用程序的所有模块
type Mods struct {
	RBAC *rbac.RBAC // RBAC模块实例
}

// Init 初始化所有模块
// ctx 上下文参数用于控制初始化过程
func (a *Mods) Init(ctx context.Context) error {
	// 初始化RBAC模块
	if err := a.RBAC.Init(ctx); err != nil {
		return err
	}

	return nil
}

// RouterPrefixes 返回API路由前缀列表
func (a *Mods) RouterPrefixes() []string {
	return []string{
		apiPrefix,
	}
}

// RegisterRouters 注册所有模块的路由
// ctx 上下文参数
// e 是gin的Engine实例，用于注册路由
func (a *Mods) RegisterRouters(ctx context.Context, e *gin.Engine) error {
	// 创建API路由组
	gAPI := e.Group(apiPrefix)
	// 创建v1版本的路由组
	v1 := gAPI.Group("v1")

	// 注册RBAC模块的v1版本路由
	if err := a.RBAC.RegisterV1Routers(ctx, v1); err != nil {
		return err
	}

	return nil
}

// Release 释放所有模块的资源
// ctx 上下文参数
func (a *Mods) Release(ctx context.Context) error {
	// 释放RBAC模块的资源
	if err := a.RBAC.Release(ctx); err != nil {
		return err
	}

	return nil
}
