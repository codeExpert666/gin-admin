// Package api 实现了日志管理的 HTTP API 接口层
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

// Logger 结构体用于日志管理，实现了日志相关的 HTTP 接口
type Logger struct {
	// LoggerBIZ 是业务逻辑层的实例，用于处理具体的日志业务逻辑
	LoggerBIZ *biz.Logger
}

// @Tags LoggerAPI
// @Security ApiKeyAuth
// @Summary Query logger list
// @Param current query int true "pagination index" default(1)
// @Param pageSize query int true "pagination size" default(10)
// @Param level query string false "log level"
// @Param traceID query string false "trace ID"
// @Param userName query string false "user name"
// @Param tag query string false "log tag"
// @Param message query string false "log message"
// @Param startTime query string false "start time"
// @Param endTime query string false "end time"
// @Success 200 {object} util.ResponseResult{data=[]schema.Logger}
// @Failure 401 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/v1/loggers [get]
// Query 处理日志查询的 HTTP GET 请求
// 支持分页查询和多个过滤条件（日志级别、追踪ID、用户名、标签、消息内容、时间范围等）
func (a *Logger) Query(c *gin.Context) {
	// 获取请求的上下文
	ctx := c.Request.Context()
	
	// 定义查询参数结构体
	var params schema.LoggerQueryParam
	
	// 解析 URL 查询参数到结构体中
	// 如果解析失败，返回错误响应
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	// 调用业务层的 Query 方法执行实际的日志查询
	// 返回查询结果和分页信息
	result, err := a.LoggerBIZ.Query(ctx, params)
	if err != nil {
		// 如果查询过程中发生错误，返回错误响应
		util.ResError(c, err)
		return
	}
	
	// 返回成功响应，包含日志数据和分页信息
	util.ResPage(c, result.Data, result.PageResult)
}
