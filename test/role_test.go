package test

import (
	"net/http"
	"testing"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestRole 测试角色管理相关的API接口
// 测试流程包括:
// 1. 创建菜单
// 2. 创建角色并关联菜单
// 3. 查询角色列表
// 4. 更新角色信息
// 5. 删除角色和菜单
func TestRole(t *testing.T) {
	// 初始化测试环境
	e := tester(t)

	// 创建菜单表单数据
	menuFormItem := schema.MenuForm{
		Code:        "role",                   // 菜单编码
		Name:        "Role management",        // 菜单名称
		Description: "Role management",        // 菜单描述
		Sequence:    8,                        // 排序值
		Type:        "page",                   // 菜单类型:页面
		Path:        "/system/role",           // 菜单路径
		Properties:  `{"icon":"role"}`,        // 菜单属性,包含图标信息
		Status:      schema.MenuStatusEnabled, // 菜单状态:启用
	}

	// 发送创建菜单的POST请求,并验证响应
	var menu schema.Menu
	e.POST(baseAPI + "/menus").WithJSON(menuFormItem).
		Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &menu})

	// 使用断言验证菜单创建结果
	assert := assert.New(t)
	assert.NotEmpty(menu.ID) // 验证生成的菜单ID不为空
	// 验证返回的菜单信息与提交的表单数据一致
	assert.Equal(menuFormItem.Code, menu.Code)
	assert.Equal(menuFormItem.Name, menu.Name)
	assert.Equal(menuFormItem.Description, menu.Description)
	assert.Equal(menuFormItem.Sequence, menu.Sequence)
	assert.Equal(menuFormItem.Type, menu.Type)
	assert.Equal(menuFormItem.Path, menu.Path)
	assert.Equal(menuFormItem.Properties, menu.Properties)
	assert.Equal(menuFormItem.Status, menu.Status)

	// 创建角色表单数据
	roleFormItem := schema.RoleForm{
		Code: "admin",         // 角色编码
		Name: "Administrator", // 角色名称
		Menus: schema.RoleMenus{ // 关联的菜单列表
			{MenuID: menu.ID}, // 关联上面创建的菜单
		},
		Description: "Administrator",          // 角色描述
		Sequence:    9,                        // 排序值
		Status:      schema.RoleStatusEnabled, // 角色状态:启用
	}

	// 发送创建角色的POST请求,并验证响应
	var role schema.Role
	e.POST(baseAPI + "/roles").WithJSON(roleFormItem).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &role})
	// 验证角色创建结果
	assert.NotEmpty(role.ID) // 验证生成的角色ID不为空
	// 验证返回的角色信息与提交的表单数据一致
	assert.Equal(roleFormItem.Code, role.Code)
	assert.Equal(roleFormItem.Name, role.Name)
	assert.Equal(roleFormItem.Description, role.Description)
	assert.Equal(roleFormItem.Sequence, role.Sequence)
	assert.Equal(roleFormItem.Status, role.Status)
	assert.Equal(len(roleFormItem.Menus), len(role.Menus))

	// 获取角色列表并验证
	var roles schema.Roles
	e.GET(baseAPI + "/roles").Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &roles})
	assert.GreaterOrEqual(len(roles), 1) // 验证至少存在一个角色

	// 更新角色信息
	newName := "Administrator 1"           // 新的角色名称
	newStatus := schema.RoleStatusDisabled // 新的角色状态:禁用
	role.Name = newName
	role.Status = newStatus
	// 发送更新请求
	e.PUT(baseAPI + "/roles/" + role.ID).WithJSON(role).Expect().Status(http.StatusOK)

	// 获取更新后的角色信息并验证
	var getRole schema.Role
	e.GET(baseAPI + "/roles/" + role.ID).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &getRole})
	assert.Equal(newName, getRole.Name)     // 验证名称已更新
	assert.Equal(newStatus, getRole.Status) // 验证状态已更新

	// 删除角色并验证
	e.DELETE(baseAPI + "/roles/" + role.ID).Expect().Status(http.StatusOK)
	// 验证角色已被删除
	e.GET(baseAPI + "/roles/" + role.ID).Expect().Status(http.StatusNotFound)

	// 删除菜单并验证
	e.DELETE(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusOK)
	// 验证菜单已被删除
	e.GET(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusNotFound)
}
