// Package schema 定义了数据库模型的结构体
package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// RoleMenu 定义了角色与菜单的关联关系
// 这是一个多对多关系的中间表，用于实现RBAC（基于角色的访问控制）
type RoleMenu struct {
	ID        string    `json:"id" gorm:"size:20;primarykey"` // 记录ID，主键
	RoleID    string    `json:"role_id" gorm:"size:20;index"` // 角色ID，关联到Role表的ID字段
	MenuID    string    `json:"menu_id" gorm:"size:20;index"` // 菜单ID，关联到Menu表的ID字段
	CreatedAt time.Time `json:"created_at" gorm:"index;"`     // 记录创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`     // 记录更新时间
}

// TableName 返回数据库表名
// 通过配置文件中的设置来格式化表名
func (a *RoleMenu) TableName() string {
	return config.C.FormatTableName("role_menu")
}

// RoleMenuQueryParam 定义了查询角色菜单关联关系的参数
type RoleMenuQueryParam struct {
	util.PaginationParam        // 嵌入分页参数结构体，包含页码、每页数量等信息
	RoleID               string `form:"-"` // 按角色ID筛选，form:"-" 表示该字段不会从表单中获取
}

// RoleMenuQueryOptions 定义了查询的选项配置
type RoleMenuQueryOptions struct {
	util.QueryOptions // 嵌入查询选项结构体，可能包含排序、筛选等通用查询选项
}

// RoleMenuQueryResult 定义了查询结果的数据结构
type RoleMenuQueryResult struct {
	Data       RoleMenus              // 查询得到的角色菜单关联记录列表
	PageResult *util.PaginationResult // 分页信息，包含总数、页码等
}

// RoleMenus 是RoleMenu结构体的切片类型
// 用于批量处理角色菜单关联记录
type RoleMenus []*RoleMenu

// RoleMenuForm 定义了创建或更新角色菜单关联时的表单结构
// 当前为空，说明可能直接使用RoleID和MenuID进行关联
type RoleMenuForm struct {
}

// Validate 验证表单数据的合法性
func (a *RoleMenuForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到RoleMenu结构体中
func (a *RoleMenuForm) FillTo(roleMenu *RoleMenu) error {
	return nil
}
