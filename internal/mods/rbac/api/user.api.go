// Package api 实现了 RBAC（基于角色的访问控制）中用户管理的 API 层
package api

import (
	// biz 包含业务逻辑层代码
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/biz"
	// schema 包含数据结构定义
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	// util 包含通用工具函数
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	// gin 是一个 HTTP web 框架
	"github.com/gin-gonic/gin"
)

// User 结构体定义了用户管理的 API 处理器
// 它封装了用户管理相关的业务逻辑接口
type User struct {
	// UserBIZ 是用户管理的业务逻辑接口
	UserBIZ *biz.User
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Query user list
// @Param current query int true "pagination index" default(1)
// @Param pageSize query int true "pagination size" default(10)
// @Param username query string false "Username for login"
// @Param name query string false "Name of user"
// @Param status query string false "Status of user (activated, freezed)"
// @Success 200 {object} util.ResponseResult{data=[]schema.User}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users [get]
// Query 处理用户列表查询请求
// 支持分页查询和多个过滤条件（用户名、姓名、状态等）
func (a *User) Query(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 定义查询参数结构体
	var params schema.UserQueryParam
	// 解析 URL 查询参数到结构体中
	// 如果解析失败，返回错误响应
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	// 调用业务层的 Query 方法执行实际的用户查询
	// 返回查询结果和分页信息
	result, err := a.UserBIZ.Query(ctx, params)
	if err != nil {
		// 如果查询过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应，包含用户数据和分页信息
	util.ResPage(c, result.Data, result.PageResult)
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Get user record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult{data=schema.User}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users/{id} [get]
// Get 处理获取单个用户信息的请求
// 根据用户 ID 获取用户的详细信息
func (a *User) Get(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 调用业务层的 Get 方法获取用户信息
	// c.Param("id") 获取 URL 路径中的 id 参数
	item, err := a.UserBIZ.Get(ctx, c.Param("id"))
	if err != nil {
		// 如果获取过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应，包含用户信息
	util.ResSuccess(c, item)
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Create user record
// @Param body body schema.UserForm true "Request body"
// @Success 200 {object} util.ResponseResult{data=schema.User}
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users [post]
// Create 处理创建新用户的请求
// 接收用户表单数据，验证后创建新用户
func (a *User) Create(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 创建用户表单对象
	item := new(schema.UserForm)
	// 解析请求体 JSON 数据到表单对象
	if err := util.ParseJSON(c, item); err != nil {
		// 如果解析失败，返回错误响应
		util.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		// 验证表单数据
		// 如果验证失败，返回错误响应
		util.ResError(c, err)
		return
	}

	// 调用业务层的 Create 方法创建用户
	result, err := a.UserBIZ.Create(ctx, item)
	if err != nil {
		// 如果创建过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应，包含创建的用户信息
	util.ResSuccess(c, result)
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Update user record by ID
// @Param id path string true "unique id"
// @Param body body schema.UserForm true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users/{id} [put]
// Update 处理更新用户信息的请求
// 根据用户 ID 更新用户信息
func (a *User) Update(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 创建用户表单对象
	item := new(schema.UserForm)
	// 解析请求体 JSON 数据到表单对象
	if err := util.ParseJSON(c, item); err != nil {
		// 如果解析失败，返回错误响应
		util.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		// 验证表单数据
		// 如果验证失败，返回错误响应
		util.ResError(c, err)
		return
	}

	// 调用业务层的 Update 方法更新用户信息
	// c.Param("id") 获取 URL 路径中的 id 参数
	err := a.UserBIZ.Update(ctx, c.Param("id"), item)
	if err != nil {
		// 如果更新过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应
	util.ResOK(c)
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Delete user record by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users/{id} [delete]
// Delete 处理删除用户的请求
// 根据用户 ID 删除指定用户
func (a *User) Delete(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 调用业务层的 Delete 方法删除用户
	// c.Param("id") 获取 URL 路径中的 id 参数
	err := a.UserBIZ.Delete(ctx, c.Param("id"))
	if err != nil {
		// 如果删除过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应
	util.ResOK(c)
}

// @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Reset user password by ID
// @Param id path string true "unique id"
// @Success 200 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/users/{id}/reset-pwd [patch]
// ResetPassword 处理重置用户密码的请求
// 根据用户 ID 将用户密码重置为系统默认密码
func (a *User) ResetPassword(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	// 调用业务层的 ResetPassword 方法重置用户密码
	// c.Param("id") 获取 URL 路径中的 id 参数
	err := a.UserBIZ.ResetPassword(ctx, c.Param("id"))
	if err != nil {
		// 如果重置过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应
	util.ResOK(c)
}
