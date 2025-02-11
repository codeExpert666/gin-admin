// Package schema 定义了数据库模型的结构体和相关方法
package schema

import (
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/crypto/hash"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/go-playground/validator/v10"
)

// 定义用户状态常量
const (
	UserStatusActivated = "activated" // 用户状态：已激活
	UserStatusFreezed   = "freezed"   // 用户状态：已冻结
)

// User 定义了RBAC（基于角色的访问控制）中用户管理的数据结构
// 使用 GORM 标签来定义数据库表结构，使用 JSON 标签来定义 JSON 序列化/反序列化的字段名
type User struct {
	ID        string    `json:"id" gorm:"size:20;primarykey;"` // 用户唯一标识符
	Username  string    `json:"username" gorm:"size:64;index"` // 用户登录名，建立索引以提高查询性能
	Name      string    `json:"name" gorm:"size:64;index"`     // 用户真实姓名
	Password  string    `json:"-" gorm:"size:64;"`             // 登录密码（加密存储）。json:"-" 表示该字段不会在JSON中显示
	Phone     string    `json:"phone" gorm:"size:32;"`         // 用户手机号
	Email     string    `json:"email" gorm:"size:128;"`        // 用户邮箱
	Remark    string    `json:"remark" gorm:"size:1024;"`      // 用户备注信息
	Status    string    `json:"status" gorm:"size:20;index"`   // 用户状态（activated-已激活, freezed-已冻结）
	CreatedAt time.Time `json:"created_at" gorm:"index;"`      // 记录创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"index;"`      // 记录更新时间
	Roles     UserRoles `json:"roles" gorm:"-"`                // 用户关联的角色列表，gorm:"-" 表示该字段不映射到数据库
}

// TableName 返回数据库表名
// 通过配置文件中的格式化方法来生成最终的表名
func (a *User) TableName() string {
	return config.C.FormatTableName("user")
}

// UserQueryParam 定义了查询用户时的参数结构
type UserQueryParam struct {
	util.PaginationParam        // 嵌入分页参数
	LikeUsername         string `form:"username"`                                   // 按用户名模糊查询
	LikeName             string `form:"name"`                                       // 按真实姓名模糊查询
	Status               string `form:"status" binding:"oneof=activated freezed '"` // 按状态查询，必须是指定值之一
}

// UserQueryOptions 定义了查询用户时的选项
type UserQueryOptions struct {
	util.QueryOptions // 嵌入查询选项
}

// UserQueryResult 定义了用户查询的结果结构
type UserQueryResult struct {
	Data       Users                  // 查询结果数据列表
	PageResult *util.PaginationResult // 分页信息
}

// Users 定义了用户对象的切片类型
type Users []*User

// ToIDs 将用户列表转换为ID列表
func (a Users) ToIDs() []string {
	var ids []string
	for _, item := range a {
		ids = append(ids, item.ID)
	}
	return ids
}

// UserForm 定义了创建或更新用户时的表单数据结构
// binding 标签用于请求参数验证
type UserForm struct {
	Username string    `json:"username" binding:"required,max=64"`                // 用户名（必填，最大64字符）
	Name     string    `json:"name" binding:"required,max=64"`                    // 真实姓名（必填，最大64字符）
	Password string    `json:"password" binding:"max=64"`                         // 密码（可选，最大64字符）
	Phone    string    `json:"phone" binding:"max=32"`                            // 手机号（可选，最大32字符）
	Email    string    `json:"email" binding:"max=128"`                           // 邮箱（可选，最大128字符）
	Remark   string    `json:"remark" binding:"max=1024"`                         // 备注（可选，最大1024字符）
	Status   string    `json:"status" binding:"required,oneof=activated freezed"` // 状态（必填，必须是activated或freezed）
	Roles    UserRoles `json:"roles" binding:"required"`                          // 用户角色（必填）
}

// Validate 验证用户表单数据
// 主要验证邮箱格式是否正确
func (a *UserForm) Validate() error {
	if a.Email != "" && validator.New().Var(a.Email, "email") != nil {
		return errors.BadRequest("", "Invalid email address")
	}
	return nil
}

// FillTo 将表单数据填充到用户对象中
// 如果提供了密码，会先进行加密处理
func (a *UserForm) FillTo(user *User) error {
	user.Username = a.Username
	user.Name = a.Name
	user.Phone = a.Phone
	user.Email = a.Email
	user.Remark = a.Remark
	user.Status = a.Status

	// 如果提供了密码，进行加密处理
	if pass := a.Password; pass != "" {
		hashPass, err := hash.GeneratePassword(pass)
		if err != nil {
			return errors.BadRequest("", "Failed to generate hash password: %s", err.Error())
		}
		user.Password = hashPass
	}

	return nil
}
