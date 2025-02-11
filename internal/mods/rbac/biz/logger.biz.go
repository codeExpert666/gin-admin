// Package biz 实现业务逻辑层，处理日志管理相关的业务操作
package biz

import (
	"context"

	// dal 包提供数据访问层接口
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	// schema 包定义了数据结构和查询参数
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	// util 包提供通用工具函数
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Logger 结构体用于日志管理
// 它封装了对日志数据的业务操作，通过 LoggerDAL 实现与数据库的交互
type Logger struct {
	// LoggerDAL 是数据访问层的实例，用于执行实际的数据库操作
	LoggerDAL *dal.Logger
}

// Query 方法用于查询日志记录
// 参数说明：
// - ctx: 上下文对象，用于传递请求上下文信息
// - params: 查询参数，包含过滤条件等
// 返回值：
// - *schema.LoggerQueryResult: 查询结果，包含日志记录列表和分页信息
// - error: 如果查询过程中发生错误，返回相应的错误信息
func (a *Logger) Query(ctx context.Context, params schema.LoggerQueryParam) (*schema.LoggerQueryResult, error) {
	// 启用分页功能
	params.Pagination = true

	// 调用数据访问层执行查询
	// 设置查询选项：按创建时间降序排序
	result, err := a.LoggerDAL.Query(ctx, params, schema.LoggerQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: []util.OrderByParam{
				{Field: "created_at", Direction: util.DESC}, // 按创建时间降序排序
			},
		},
	})
	if err != nil {
		return nil, err // 如果发生错误，返回错误信息
	}
	return result, nil // 返回查询结果
}
