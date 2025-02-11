// Package schema 定义了数据库模型的结构体和相关方法
package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// 角色状态常量定义
const (
	RoleStatusEnabled  = "enabled"  // 启用状态
	RoleStatusDisabled = "disabled" // 禁用状态

	RoleResultTypeSelect = "select" // 查询结果类型：选择模式
)

// Role 定义了RBAC（基于角色的访问控制）中的角色数据结构
// 使用 GORM 标签来定义数据库表结构
type Role struct {
	ID          string    `json:"id" gorm:"size:20;primarykey;"` // 角色ID，主键，限制长度为20
	Code        string    `json:"code" gorm:"size:32;index;"`    // 角色代码，唯一标识符，建立索引
	Name        string    `json:"name" gorm:"size:128;index"`    // 角色名称，用于显示，建立索引
	Description string    `json:"description" gorm:"size:1024"`  // 角色描述，最大长度1024
	Sequence    int       `json:"sequence" gorm:"index"`         // 排序序号，用于自定义角色显示顺序
	Status      string    `json:"status" gorm:"size:20;index"`   // 角色状态（启用/禁用），建立索引
	CreatedAt   time.Time `json:"created_at" gorm:"index;"`      // 创建时间，建立索引
	UpdatedAt   time.Time `json:"updated_at" gorm:"index;"`      // 更新时间，建立索引
	Menus       RoleMenus `json:"menus" gorm:"-"`                // 角色关联的菜单列表，使用 gorm:"-" 表示该字段不映射到数据库
}

// TableName 返回数据库表名
// 通过配置文件中的格式化方法来生成最终的表名
func (a *Role) TableName() string {
	return config.C.FormatTableName("role")
}

// RoleQueryParam 定义了角色查询的参数结构
type RoleQueryParam struct {
	util.PaginationParam            // 嵌入分页参数
	LikeName             string     `form:"name"`                                       // 按角色名称模糊查询
	Status               string     `form:"status" binding:"oneof=disabled enabled ''"` // 按状态查询，限制只能是disabled或enabled
	ResultType           string     `form:"resultType"`                                 // 结果类型
	InIDs                []string   `form:"-"`                                          // 根据ID列表查询，form:"-" 表示不从请求参数中绑定
	GtUpdatedAt          *time.Time `form:"-"`                                          // 查询更新时间大于指定时间的记录
}

// RoleQueryOptions 定义了角色查询的选项
type RoleQueryOptions struct {
	util.QueryOptions // 嵌入查询选项基础结构
}

// RoleQueryResult 定义了角色查询的结果结构
type RoleQueryResult struct {
	Data       Roles                  // 角色数据列表
	PageResult *util.PaginationResult // 分页结果信息
}

// Roles 定义了角色列表类型
type Roles []*Role

// RoleForm 定义了创建/更新角色时的表单数据结构
type RoleForm struct {
	Code        string    `json:"code" binding:"required,max=32"`                   // 角色代码，必填，最大长度32
	Name        string    `json:"name" binding:"required,max=128"`                  // 角色名称，必填，最大长度128
	Description string    `json:"description"`                                      // 角色描述，选填
	Sequence    int       `json:"sequence"`                                         // 排序序号，选填
	Status      string    `json:"status" binding:"required,oneof=disabled enabled"` // 状态，必填，只能是disabled或enabled
	Menus       RoleMenus `json:"menus"`                                            // 角色关联的菜单列表
}

// Validate 验证表单数据
// 目前为空实现，可以根据需要添加自定义验证逻辑
func (a *RoleForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到Role对象中
func (a *RoleForm) FillTo(role *Role) error {
	role.Code = a.Code
	role.Name = a.Name
	role.Description = a.Description
	role.Sequence = a.Sequence
	role.Status = a.Status
	return nil
}
