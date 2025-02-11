package schema

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// 定义菜单状态常量
const (
	MenuStatusDisabled = "disabled" // 菜单禁用状态
	MenuStatusEnabled  = "enabled"  // 菜单启用状态
)

// 定义菜单排序参数
var (
	MenusOrderParams = []util.OrderByParam{
		{Field: "sequence", Direction: util.DESC},   // 按序号降序排序
		{Field: "created_at", Direction: util.DESC}, // 按创建时间降序排序
	}
)

// Menu 结构体定义了RBAC（基于角色的访问控制）中的菜单管理
type Menu struct {
	ID          string        `json:"id" gorm:"size:20;primarykey;"`      // 唯一标识符
	Code        string        `json:"code" gorm:"size:32;index;"`         // 菜单编码（每个层级都是唯一的）
	Name        string        `json:"name" gorm:"size:128;index"`         // 菜单显示名称
	Description string        `json:"description" gorm:"size:1024"`       // 菜单描述信息
	Sequence    int           `json:"sequence" gorm:"index;"`             // 排序序号（降序排序）
	Type        string        `json:"type" gorm:"size:20;index"`          // 菜单类型（页面、按钮）
	Path        string        `json:"path" gorm:"size:255;"`              // 菜单访问路径
	Properties  string        `json:"properties" gorm:"type:text;"`       // 菜单属性（JSON格式）
	Status      string        `json:"status" gorm:"size:20;index"`        // 菜单状态（启用、禁用）
	ParentID    string        `json:"parent_id" gorm:"size:20;index;"`    // 父级ID（关联Menu.ID）
	ParentPath  string        `json:"parent_path" gorm:"size:255;index;"` // 父级路径（用.分隔）
	Children    *Menus        `json:"children" gorm:"-"`                  // 子菜单列表
	CreatedAt   time.Time     `json:"created_at" gorm:"index;"`           // 创建时间
	UpdatedAt   time.Time     `json:"updated_at" gorm:"index;"`           // 更新时间
	Resources   MenuResources `json:"resources" gorm:"-"`                 // 菜单关联的资源列表
}

// TableName 返回数据库表名
func (a *Menu) TableName() string {
	return config.C.FormatTableName("menu")
}

// MenuQueryParam 定义了菜单查询的参数结构
type MenuQueryParam struct {
	util.PaginationParam          // 继承分页参数
	CodePath             string   `form:"code"`             // 编码路径（格式：xxx.xxx.xxx）
	LikeName             string   `form:"name"`             // 菜单名称（模糊查询）
	IncludeResources     bool     `form:"includeResources"` // 是否包含资源信息
	InIDs                []string `form:"-"`                // 指定的菜单ID列表
	Status               string   `form:"-"`                // 菜单状态
	ParentID             string   `form:"-"`                // 父级ID
	ParentPathPrefix     string   `form:"-"`                // 父级路径前缀
	UserID               string   `form:"-"`                // 用户ID
	RoleID               string   `form:"-"`                // 角色ID
}

// MenuQueryOptions 定义查询选项
type MenuQueryOptions struct {
	util.QueryOptions // 继承查询选项
}

// MenuQueryResult 定义查询结果
type MenuQueryResult struct {
	Data       Menus                  // 菜单数据列表
	PageResult *util.PaginationResult // 分页结果信息
}

// Menus 定义菜单切片类型
type Menus []*Menu

// 实现sort.Interface接口的方法，用于菜单排序
func (a Menus) Len() int { return len(a) }
func (a Menus) Less(i, j int) bool {
	if a[i].Sequence == a[j].Sequence {
		return a[i].CreatedAt.Unix() > a[j].CreatedAt.Unix()
	}
	return a[i].Sequence > a[j].Sequence
}
func (a Menus) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ToMap 将菜单列表转换为以ID为键的map
func (a Menus) ToMap() map[string]*Menu {
	m := make(map[string]*Menu)
	for _, item := range a {
		m[item.ID] = item
	}
	return m
}

// SplitParentIDs 提取所有父级ID
func (a Menus) SplitParentIDs() []string {
	parentIDs := make([]string, 0, len(a))
	idMapper := make(map[string]struct{})
	for _, item := range a {
		if _, ok := idMapper[item.ID]; ok {
			continue
		}
		idMapper[item.ID] = struct{}{}
		if pp := item.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, util.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := idMapper[pid]; ok {
					continue
				}
				parentIDs = append(parentIDs, pid)
				idMapper[pid] = struct{}{}
			}
		}
	}
	return parentIDs
}

// ToTree 将菜单列表转换为树形结构
func (a Menus) ToTree() Menus {
	var list Menus
	m := a.ToMap()
	for _, item := range a {
		if item.ParentID == "" {
			list = append(list, item)
			continue
		}
		if parent, ok := m[item.ParentID]; ok {
			if parent.Children == nil {
				children := Menus{item}
				parent.Children = &children
				continue
			}
			*parent.Children = append(*parent.Children, item)
		}
	}
	return list
}

// MenuForm 定义创建或更新菜单时的表单结构
type MenuForm struct {
	Code        string        `json:"code" binding:"required,max=32"`                   // 菜单编码（必填，最大32字符）
	Name        string        `json:"name" binding:"required,max=128"`                  // 菜单名称（必填，最大128字符）
	Description string        `json:"description"`                                      // 描述信息
	Sequence    int           `json:"sequence"`                                         // 排序序号
	Type        string        `json:"type" binding:"required,oneof=page button"`        // 菜单类型（必填，只能是page或button）
	Path        string        `json:"path"`                                             // 访问路径
	Properties  string        `json:"properties"`                                       // 扩展属性（JSON格式）
	Status      string        `json:"status" binding:"required,oneof=disabled enabled"` // 状态（必填，只能是disabled或enabled）
	ParentID    string        `json:"parent_id"`                                        // 父级ID
	Resources   MenuResources `json:"resources"`                                        // 关联的资源列表
}

// Validate 验证表单数据
func (a *MenuForm) Validate() error {
	if v := a.Properties; v != "" {
		if !json.Valid([]byte(v)) {
			return errors.BadRequest("", "invalid properties")
		}
	}
	return nil
}

// FillTo 将表单数据填充到菜单对象中
func (a *MenuForm) FillTo(menu *Menu) error {
	menu.Code = a.Code
	menu.Name = a.Name
	menu.Description = a.Description
	menu.Sequence = a.Sequence
	menu.Type = a.Type
	menu.Path = a.Path
	menu.Properties = a.Properties
	menu.Status = a.Status
	menu.ParentID = a.ParentID
	return nil
}
