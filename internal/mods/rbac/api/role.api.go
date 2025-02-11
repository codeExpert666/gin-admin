// Package api 实现了角色管理的 HTTP API 接口层
package api

import (
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/biz"    // 导入业务逻辑层
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema" // 导入数据结构定义
	"github.com/LyricTian/gin-admin/v10/pkg/util"                  // 导入工具包
	"github.com/gin-gonic/gin"                                     // 导入 Gin Web 框架
)

// Role 结构体定义了角色管理的 API 处理器
// 它封装了角色管理相关的业务逻辑接口
type Role struct {
	RoleBIZ *biz.Role // 角色管理的业务逻辑接口
}

// Query 处理角色列表查询请求
// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary 查询角色列表
// @Param current query int true "分页索引" default(1)
// @Param pageSize query int true "分页大小" default(10)
// @Param name query string false "角色显示名称"
// @Param status query string false "角色状态(disabled-禁用, enabled-启用)"
// @Success 200 {object} util.ResponseResult{data=[]schema.Role}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles [get]
func (a *Role) Query(c *gin.Context) {
	ctx := c.Request.Context()                          // 获取请求上下文
	var params schema.RoleQueryParam                    // 定义查询参数结构体
	if err := util.ParseQuery(c, &params); err != nil { // 解析查询参数
		util.ResError(c, err) // 返回错误响应
		return
	}

	result, err := a.RoleBIZ.Query(ctx, params) // 调用业务层执行查询
	if err != nil {
		util.ResError(c, err) // 返回错误响应
		return
	}
	util.ResPage(c, result.Data, result.PageResult) // 返回分页数据
}

// Get 处理获取指定角色信息的请求
// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary 根据ID获取角色信息
// @Param id path string true "角色ID"
// @Success 200 {object} util.ResponseResult{data=schema.Role}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [get]
func (a *Role) Get(c *gin.Context) {
	ctx := c.Request.Context()                     // 获取请求上下文
	item, err := a.RoleBIZ.Get(ctx, c.Param("id")) // 调用业务层获取角色信息
	if err != nil {
		util.ResError(c, err) // 返回错误响应
		return
	}
	util.ResSuccess(c, item) // 返回成功响应
}

// Create 处理创建新角色的请求
// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary 创建新角色
// @Param body body schema.RoleForm true "请求体"
// @Success 200 {object} util.ResponseResult{data=schema.Role}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles [post]
func (a *Role) Create(c *gin.Context) {
	ctx := c.Request.Context()                      // 获取请求上下文
	item := new(schema.RoleForm)                    // 创建角色表单对象
	if err := util.ParseJSON(c, item); err != nil { // 解析请求JSON数据
		util.ResError(c, err) // 返回错误响应
		return
	} else if err := item.Validate(); err != nil { // 验证表单数据
		util.ResError(c, err) // 返回验证错误
		return
	}

	result, err := a.RoleBIZ.Create(ctx, item) // 调用业务层创建角色
	if err != nil {
		util.ResError(c, err) // 返回错误响应
		return
	}
	util.ResSuccess(c, result) // 返回创建成功的角色信息
}

// Update 处理更新角色信息的请求
// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary 更新指定角色信息
// @Param id path string true "角色ID"
// @Param body body schema.RoleForm true "请求体"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [put]
func (a *Role) Update(c *gin.Context) {
	ctx := c.Request.Context()                      // 获取请求上下文
	item := new(schema.RoleForm)                    // 创建角色表单对象
	if err := util.ParseJSON(c, item); err != nil { // 解析请求JSON数据
		util.ResError(c, err) // 返回错误响应
		return
	} else if err := item.Validate(); err != nil { // 验证表单数据
		util.ResError(c, err) // 返回验证错误
		return
	}

	err := a.RoleBIZ.Update(ctx, c.Param("id"), item) // 调用业务层更新角色
	if err != nil {
		util.ResError(c, err) // 返回错误响应
		return
	}
	util.ResOK(c) // 返回更新成功响应
}

// Delete 处理删除角色的请求
// @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary 删除指定角色
// @Param id path string true "角色ID"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/roles/{id} [delete]
func (a *Role) Delete(c *gin.Context) {
	ctx := c.Request.Context()                  // 获取请求上下文
	err := a.RoleBIZ.Delete(ctx, c.Param("id")) // 调用业务层删除角色
	if err != nil {
		util.ResError(c, err) // 返回错误响应
		return
	}
	util.ResOK(c) // 返回删除成功响应
}
