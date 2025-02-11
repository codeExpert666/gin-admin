package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// UserRole 定义了用户和角色之间的关联关系
// 这是一个多对多关系的中间表，用于实现RBAC（基于角色的访问控制）
type UserRole struct {
	ID        string    `json:"id" gorm:"size:20;primarykey"`           // 唯一标识符，主键
	UserID    string    `json:"user_id" gorm:"size:20;index"`           // 关联的用户ID，建立索引以提高查询性能
	RoleID    string    `json:"role_id" gorm:"size:20;index"`           // 关联的角色ID，建立索引以提高查询性能
	CreatedAt time.Time `json:"created_at" gorm:"index;"`               // 记录创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`               // 记录更新时间
	RoleName  string    `json:"role_name" gorm:"<-:false;-:migration;"` // 角色名称（只读字段，不参与数据库迁移）
}

// TableName 返回数据库表名
// 通过配置文件中的设置来格式化最终的表名
func (a *UserRole) TableName() string {
	return config.C.FormatTableName("user_role")
}

// UserRoleQueryParam 定义查询用户角色关系时的参数结构
type UserRoleQueryParam struct {
	util.PaginationParam          // 嵌入分页参数，继承分页相关的功能
	InUserIDs            []string `form:"-"` // 按用户ID列表进行过滤
	UserID               string   `form:"-"` // 按单个用户ID进行过滤
	RoleID               string   `form:"-"` // 按角色ID进行过滤
}

// UserRoleQueryOptions 定义查询时的额外选项
type UserRoleQueryOptions struct {
	util.QueryOptions      // 嵌入基础查询选项
	JoinRole          bool // 是否关联查询角色表的数据
}

// UserRoleQueryResult 定义查询结果的数据结构
type UserRoleQueryResult struct {
	Data       UserRoles              // 查询得到的用户角色关系数据
	PageResult *util.PaginationResult // 分页信息
}

// UserRoles 是UserRole的切片类型，用于批量处理用户角色关系
type UserRoles []*UserRole

// ToUserIDMap 将用户角色关系转换为以用户ID为键的映射
// 返回一个map，其中键是用户ID，值是该用户拥有的所有角色关系
func (a UserRoles) ToUserIDMap() map[string]UserRoles {
	m := make(map[string]UserRoles)
	for _, userRole := range a {
		m[userRole.UserID] = append(m[userRole.UserID], userRole)
	}
	return m
}

// ToRoleIDs 提取所有角色ID并返回一个切片
func (a UserRoles) ToRoleIDs() []string {
	var ids []string
	for _, item := range a {
		ids = append(ids, item.RoleID)
	}
	return ids
}

// UserRoleForm 定义创建或更新用户角色关系时的表单结构
type UserRoleForm struct {
}

// Validate 验证表单数据的合法性
func (a *UserRoleForm) Validate() error {
	return nil
}

// FillTo 将表单数据填充到UserRole实体中
func (a *UserRoleForm) FillTo(userRole *UserRole) error {
	return nil
}
