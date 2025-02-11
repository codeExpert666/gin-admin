package api

import (
	// 导入业务逻辑层
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/biz"
	// 导入数据结构定义
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	// 导入工具包
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	// 导入 Gin Web 框架
	"github.com/gin-gonic/gin"
)

// Login 登录模块的 API 处理结构体
type Login struct {
	// LoginBIZ 是登录模块的业务逻辑层接口
	LoginBIZ *biz.Login
}

// GetCaptcha 获取验证码接口
// @Tags LoginAPI
// @Summary Get captcha ID
// @Success 200 {object} util.ResponseResult{data=schema.Captcha}
// @Router /api/v1/captcha/id [get]
func (a *Login) GetCaptcha(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层获取验证码
	data, err := a.LoginBIZ.GetCaptcha(ctx)
	if err != nil {
		// 如果发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	// 返回成功响应，包含验证码数据
	util.ResSuccess(c, data)
}

// ResponseCaptcha 响应验证码图片接口
// @Tags LoginAPI
// @Summary Response captcha image
// @Param id query string true "Captcha ID"
// @Param reload query number false "Reload captcha image (reload=1)"
// @Produce image/png
// @Success 200 "Captcha image"
// @Failure 404 {object} util.ResponseResult
// @Router /api/v1/captcha/image [get]
func (a *Login) ResponseCaptcha(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层生成验证码图片
	// c.Writer: 响应写入器
	// c.Query("id"): 获取验证码ID
	// c.Query("reload") == "1": 判断是否需要重新生成验证码
	err := a.LoginBIZ.ResponseCaptcha(ctx, c.Writer, c.Query("id"), c.Query("reload") == "1")
	if err != nil {
		// 如果发生错误，返回错误响应
		util.ResError(c, err)
	}
}

// Login 用户登录接口
// @Tags LoginAPI
// @Summary Login system with username and password
// @Param body body schema.LoginForm true "Request body"
// @Success 200 {object} util.ResponseResult{data=schema.LoginToken}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/login [post]
func (a *Login) Login(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 创建登录表单对象
	item := new(schema.LoginForm)
	// 解析 JSON 请求体到登录表单对象
	if err := util.ParseJSON(c, item); err != nil {
		// 如果解析失败，返回错误响应
		util.ResError(c, err)
		return
	}

	// 调用业务层处理登录逻辑
	// item.Trim(): 处理表单数据（如去除空格等）
	data, err := a.LoginBIZ.Login(ctx, item.Trim())
	if err != nil {
		// 如果登录失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 登录成功，返回令牌信息
	util.ResSuccess(c, data)
}

// Logout 用户登出接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Logout system
// @Success 200 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/logout [post]
func (a *Login) Logout(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层处理登出逻辑
	err := a.LoginBIZ.Logout(ctx)
	if err != nil {
		// 如果登出失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 登出成功，返回成功响应
	util.ResOK(c)
}

// RefreshToken 刷新访问令牌接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Refresh current access token
// @Success 200 {object} util.ResponseResult{data=schema.LoginToken}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/refresh-token [post]
func (a *Login) RefreshToken(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层刷新令牌
	data, err := a.LoginBIZ.RefreshToken(ctx)
	if err != nil {
		// 如果刷新失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 刷新成功，返回新的令牌信息
	util.ResSuccess(c, data)
}

// GetUserInfo 获取当前用户信息接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Get current user info
// @Success 200 {object} util.ResponseResult{data=schema.User}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/user [get]
func (a *Login) GetUserInfo(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层获取用户信息
	data, err := a.LoginBIZ.GetUserInfo(ctx)
	if err != nil {
		// 如果获取失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 获取成功，返回用户信息
	util.ResSuccess(c, data)
}

// UpdatePassword 修改当前用户密码接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Change current user password
// @Param body body schema.UpdateLoginPassword true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/password [put]
func (a *Login) UpdatePassword(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 创建更新密码表单对象
	item := new(schema.UpdateLoginPassword)
	// 解析 JSON 请求体到表单对象
	if err := util.ParseJSON(c, item); err != nil {
		// 如果解析失败，返回错误响应
		util.ResError(c, err)
		return
	}

	// 调用业务层更新密码
	err := a.LoginBIZ.UpdatePassword(ctx, item)
	if err != nil {
		// 如果更新失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 更新成功，返回成功响应
	util.ResOK(c)
}

// QueryMenus 查询当前用户菜单接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Query current user menus based on the current user role
// @Success 200 {object} util.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/menus [get]
func (a *Login) QueryMenus(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 调用业务层查询用户菜单
	data, err := a.LoginBIZ.QueryMenus(ctx)
	if err != nil {
		// 如果查询失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 查询成功，返回菜单数据
	util.ResSuccess(c, data)
}

// UpdateUser 更新当前用户信息接口
// @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Update current user info
// @Param body body schema.UpdateCurrentUser true "Request body"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/current/user [put]
func (a *Login) UpdateUser(c *gin.Context) {
	// 获取请求上下文
	ctx := c.Request.Context()
	// 创建更新用户信息表单对象
	item := new(schema.UpdateCurrentUser)
	// 解析 JSON 请求体到表单对象
	if err := util.ParseJSON(c, item); err != nil {
		// 如果解析失败，返回错误响应
		util.ResError(c, err)
		return
	}

	// 调用业务层更新用户信息
	err := a.LoginBIZ.UpdateUser(ctx, item)
	if err != nil {
		// 如果更新失败，返回错误响应
		util.ResError(c, err)
		return
	}
	// 更新成功，返回成功响应
	util.ResOK(c)
}
