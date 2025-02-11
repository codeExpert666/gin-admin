// Package schema 定义了数据库模型的结构体
package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// Logger 结构体用于日志管理
// 这是一个数据库模型，用于存储系统的日志信息
// 使用 GORM 标签来定义数据库字段的属性
type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`           // ID：日志的唯一标识符
	Level     string    `gorm:"size:20;index;" json:"level"`             // Level：日志级别（如：info、error、debug等）
	TraceID   string    `gorm:"size:64;index;" json:"trace_id"`          // TraceID：用于分布式追踪的ID
	UserID    string    `gorm:"size:20;index;" json:"user_id"`           // UserID：记录操作用户的ID
	Tag       string    `gorm:"size:32;index;" json:"tag"`               // Tag：日志标签，用于分类
	Message   string    `gorm:"size:1024;" json:"message"`               // Message：日志的具体内容
	Stack     string    `gorm:"type:text;" json:"stack"`                 // Stack：错误堆栈信息
	Data      string    `gorm:"type:text;" json:"data"`                  // Data：额外的日志数据，JSON格式
	CreatedAt time.Time `gorm:"index;" json:"created_at"`                // CreatedAt：日志创建时间
	LoginName string    `json:"login_name" gorm:"<-:false;-:migration;"` // LoginName：用户登录名（不存储在数据库中）
	UserName  string    `json:"user_name" gorm:"<-:false;-:migration;"`  // UserName：用户名称（不存储在数据库中）
}

// TableName 方法定义了该结构体在数据库中对应的表名
// 通过配置文件中的设置来格式化表名
func (a *Logger) TableName() string {
	return config.C.FormatTableName("logger")
}

// LoggerQueryParam 定义了查询日志时的参数结构
// 继承了分页参数，并定义了多个查询条件
type LoggerQueryParam struct {
	util.PaginationParam        // 嵌入分页参数结构体
	Level                string `form:"level"`     // 按日志级别筛选
	TraceID              string `form:"traceID"`   // 按追踪ID筛选
	LikeUserName         string `form:"userName"`  // 按用户名模糊查询
	Tag                  string `form:"tag"`       // 按标签筛选
	LikeMessage          string `form:"message"`   // 按日志内容模糊查询
	StartTime            string `form:"startTime"` // 查询的开始时间
	EndTime              string `form:"endTime"`   // 查询的结束时间
}

// LoggerQueryOptions 定义了日志查询的选项
// 继承了基础查询选项结构体
type LoggerQueryOptions struct {
	util.QueryOptions // 嵌入查询选项结构体
}

// LoggerQueryResult 定义了日志查询的结果结构
type LoggerQueryResult struct {
	Data       Loggers                // 查询得到的日志数据列表
	PageResult *util.PaginationResult // 分页信息
}

// Loggers 是Logger结构体的切片类型
// 用于批量处理日志数据
type Loggers []*Logger
