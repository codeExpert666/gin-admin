// Package dal 实现数据访问层（Data Access Layer）的功能
package dal

import (
	"context"
	"fmt"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"gorm.io/gorm"
)

// GetLoggerDB 获取日志存储实例
// ctx: 上下文对象，用于传递请求相关的信息
// defDB: 默认的数据库连接
// 返回: 配置好的用于操作 Logger 表的 GORM DB 实例
func GetLoggerDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(new(schema.Logger))
}

// Logger 结构体定义了日志管理相关的数据库操作
type Logger struct {
	DB *gorm.DB // 数据库连接实例
}

// Query 方法用于根据查询参数从数据库中检索日志记录
// 参数说明:
// - ctx: 上下文对象
// - params: 查询参数，包含过滤条件
// - opts: 可选的查询选项
// 返回:
// - *schema.LoggerQueryResult: 查询结果，包含分页信息和日志数据
// - error: 可能发生的错误
func (a *Logger) Query(ctx context.Context, params schema.LoggerQueryParam, opts ...schema.LoggerQueryOptions) (*schema.LoggerQueryResult, error) {
	// 处理可选的查询选项
	var opt schema.LoggerQueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 构建基础查询
	// 使用表别名 'a' 来指代 Logger 表
	db := a.DB.Table(fmt.Sprintf("%s AS a", new(schema.Logger).TableName()))
	// 通过左连接关联用户表，以获取用户相关信息
	db = db.Joins(fmt.Sprintf("left join %s b on a.user_id=b.id", new(schema.User).TableName()))
	// 选择需要返回的字段，包括日志表的所有字段和用户表的特定字段
	db = db.Select("a.*,b.name as user_name,b.username as login_name")

	// 根据查询参数添加过滤条件
	if v := params.Level; v != "" {
		db = db.Where("a.level = ?", v) // 按日志级别过滤
	}
	if v := params.LikeMessage; len(v) > 0 {
		db = db.Where("a.message LIKE ?", "%"+v+"%") // 按日志消息模糊匹配
	}
	if v := params.TraceID; v != "" {
		db = db.Where("a.trace_id = ?", v) // 按追踪ID过滤
	}
	if v := params.LikeUserName; v != "" {
		db = db.Where("b.username LIKE ?", "%"+v+"%") // 按用户名模糊匹配
	}
	if v := params.Tag; v != "" {
		db = db.Where("a.tag = ?", v) // 按标签过滤
	}
	if start, end := params.StartTime, params.EndTime; start != "" && end != "" {
		db = db.Where("a.created_at BETWEEN ? AND ?", start, end) // 按时间范围过滤
	}

	// 执行分页查询
	var list schema.Loggers
	pageResult, err := util.WrapPageQuery(ctx, db, params.PaginationParam, opt.QueryOptions, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// 构建并返回查询结果
	queryResult := &schema.LoggerQueryResult{
		PageResult: pageResult,
		Data:       list,
	}
	return queryResult, nil
}
