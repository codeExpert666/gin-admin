// Package schema 定义了数据库模型的结构体和相关操作
package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// MenuResource 定义了菜单资源管理的数据结构
// 这个结构体用于RBAC（基于角色的访问控制）中管理API资源
// 它将菜单与具体的API路径关联起来
type MenuResource struct {
	ID        string    `json:"id" gorm:"size:20;primarykey"` // 记录的唯一标识符
	MenuID    string    `json:"menu_id" gorm:"size:20;index"` // 关联的菜单ID，外键关联到Menu表
	Method    string    `json:"method" gorm:"size:20;"`       // HTTP请求方法（如GET、POST、PUT、DELETE等）
	Path      string    `json:"path" gorm:"size:255;"`        // API请求路径（例如：/api/v1/users/:id）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`     // 记录创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`     // 记录最后更新时间
}

// TableName 返回数据库表名
// 通过配置文件中的设置格式化表名
func (a *MenuResource) TableName() string {
	return config.C.FormatTableName("menu_resource")
}

// MenuResourceQueryParam 定义了查询菜单资源时的参数结构
// 继承了分页参数，支持分页查询
type MenuResourceQueryParam struct {
	util.PaginationParam          // 嵌入分页参数结构体，提供分页功能
	MenuID               string   `form:"-"` // 按菜单ID查询，不从表单获取
	MenuIDs              []string `form:"-"` // 按多个菜单ID查询，不从表单获取
}

// MenuResourceQueryOptions 定义了查询时的选项
// 包含了一些通用的查询选项设置
type MenuResourceQueryOptions struct {
	util.QueryOptions // 嵌入查询选项结构体，提供基础的查询选项
}

// MenuResourceQueryResult 定义了查询结果的数据结构
// 包含查询得到的数据列表和分页信息
type MenuResourceQueryResult struct {
	Data       MenuResources          // 查询结果数据列表
	PageResult *util.PaginationResult // 分页信息
}

// MenuResources 是MenuResource的切片类型
// 用于批量处理菜单资源数据
type MenuResources []*MenuResource

// MenuResourceForm 定义了创建或更新菜单资源时的表单结构
// 目前为空，可以根据需求添加字段
type MenuResourceForm struct {
}

// Validate 验证表单数据的合法性
// 目前为空实现，可以根据需求添加验证规则
func (a *MenuResourceForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到MenuResource结构体中
// 目前为空实现，可以根据需求添加数据转换逻辑
func (a *MenuResourceForm) FillTo(menuResource *MenuResource) error {
	return nil
}
