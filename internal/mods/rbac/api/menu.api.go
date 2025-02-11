// package api 实现了RBAC（基于角色的访问控制）中菜单管理的API层
package api

import (
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/biz"    // 导入业务逻辑层
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema" // 导入数据模型定义
	"github.com/LyricTian/gin-admin/v10/pkg/util"                  // 导入工具包
	"github.com/gin-gonic/gin"                                     // 导入Gin Web框架
)

// Menu 结构体定义了菜单管理的API处理器
// 它封装了菜单管理相关的业务逻辑操作
type Menu struct {
	MenuBIZ *biz.Menu // MenuBIZ 是菜单管理的业务逻辑实现
}

// Query 处理菜单查询请求
// @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary 查询菜单树形数据
// @Param code query string false "菜单代码路径（格式如：xxx.xxx.xxx）"
// @Param name query string false "菜单名称"
// @Param includeResources query bool false "是否包含菜单资源"
// @Success 200 {object} util.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus [get]
func (a *Menu) Query(c *gin.Context) {
	ctx := c.Request.Context()                          // 获取请求上下文
	var params schema.MenuQueryParam                    // 定义查询参数结构体
	if err := util.ParseQuery(c, &params); err != nil { // 解析查询参数
		util.ResError(c, err) // 如果解析失败，返回错误响应
		return
	}

	result, err := a.MenuBIZ.Query(ctx, params) // 调用业务层执行查询
	if err != nil {
		util.ResError(c, err) // 如果查询失败，返回错误响应
		return
	}
	util.ResPage(c, result.Data, result.PageResult) // 返回分页查询结果
}

// Get 处理获取单个菜单记录的请求
// @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary 根据ID获取菜单记录
// @Param id path string true "唯一标识"
// @Success 200 {object} util.ResponseResult{data=schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [get]
func (a *Menu) Get(c *gin.Context) {
	ctx := c.Request.Context()                     // 获取请求上下文
	item, err := a.MenuBIZ.Get(ctx, c.Param("id")) // 调用业务层获取菜单记录
	if err != nil {
		util.ResError(c, err) // 如果获取失败，返回错误响应
		return
	}
	util.ResSuccess(c, item) // 返回成功响应和菜单数据
}

// Create 处理创建菜单的请求
// @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary 创建菜单记录
// @Param body body schema.MenuForm true "请求体"
// @Success 200 {object} util.ResponseResult{data=schema.Menu}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus [post]
func (a *Menu) Create(c *gin.Context) {
	ctx := c.Request.Context()                      // 获取请求上下文
	item := new(schema.MenuForm)                    // 创建菜单表单对象
	if err := util.ParseJSON(c, item); err != nil { // 解析请求JSON数据
		util.ResError(c, err) // 如果解析失败，返回错误响应
		return
	} else if err := item.Validate(); err != nil { // 验证表单数据
		util.ResError(c, err) // 如果验证失败，返回错误响应
		return
	}

	result, err := a.MenuBIZ.Create(ctx, item) // 调用业务层创建菜单
	if err != nil {
		util.ResError(c, err) // 如果创建失败，返回错误响应
		return
	}
	util.ResSuccess(c, result) // 返回成功响应和创建的菜单数据
}

// Update 处理更新菜单的请求
// @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary 根据ID更新菜单记录
// @Param id path string true "唯一标识"
// @Param body body schema.MenuForm true "请求体"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [put]
func (a *Menu) Update(c *gin.Context) {
	ctx := c.Request.Context()                      // 获取请求上下文
	item := new(schema.MenuForm)                    // 创建菜单表单对象
	if err := util.ParseJSON(c, item); err != nil { // 解析请求JSON数据
		util.ResError(c, err) // 如果解析失败，返回错误响应
		return
	} else if err := item.Validate(); err != nil { // 验证表单数据
		util.ResError(c, err) // 如果验证失败，返回错误响应
		return
	}

	err := a.MenuBIZ.Update(ctx, c.Param("id"), item) // 调用业务层更新菜单
	if err != nil {
		util.ResError(c, err) // 如果更新失败，返回错误响应
		return
	}
	util.ResOK(c) // 返回成功响应
}

// Delete 处理删除菜单的请求
// @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary 根据ID删除菜单记录
// @Param id path string true "唯一标识"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/menus/{id} [delete]
func (a *Menu) Delete(c *gin.Context) {
	ctx := c.Request.Context()                  // 获取请求上下文
	err := a.MenuBIZ.Delete(ctx, c.Param("id")) // 调用业务层删除菜单
	if err != nil {
		util.ResError(c, err) // 如果删除失败，返回错误响应
		return
	}
	util.ResOK(c) // 返回成功响应
}
