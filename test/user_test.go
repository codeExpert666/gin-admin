// package test 包含系统的集成测试用例
package test

import (
	"net/http"
	"testing"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/crypto/hash"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestUser 测试用户管理相关的API接口
// 测试流程:创建菜单 -> 创建角色 -> 创建用户 -> 查询用户 -> 更新用户 -> 删除用户/角色/菜单
func TestUser(t *testing.T) {
	// 初始化测试环境
	e := tester(t)

	// 第一步:创建菜单
	// 定义菜单表单数据
	menuFormItem := schema.MenuForm{
		Code:        "user",                   // 菜单编码
		Name:        "User management",        // 菜单名称
		Description: "User management",        // 菜单描述
		Sequence:    7,                        // 排序值
		Type:        "page",                   // 菜单类型:页面
		Path:        "/system/user",           // 菜单路径
		Properties:  `{"icon":"user"}`,        // 菜单属性(图标)
		Status:      schema.MenuStatusEnabled, // 菜单状态:启用
	}

	// 发送创建菜单的POST请求并验证响应
	var menu schema.Menu
	e.POST(baseAPI + "/menus").WithJSON(menuFormItem).
		Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &menu})

	// 使用assert进行断言验证菜单创建结果
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

	// 第二步:创建角色
	// 定义角色表单数据
	roleFormItem := schema.RoleForm{
		Code: "user",   // 角色编码
		Name: "Normal", // 角色名称
		Menus: schema.RoleMenus{ // 角色关联的菜单
			{MenuID: menu.ID}, // 关联上面创建的菜单
		},
		Description: "Normal",                 // 角色描述
		Sequence:    8,                        // 排序值
		Status:      schema.RoleStatusEnabled, // 角色状态:启用
	}

	// 发送创建角色的POST请求并验证响应
	var role schema.Role
	e.POST(baseAPI + "/roles").WithJSON(roleFormItem).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &role})
	// 验证角色创建结果
	assert.NotEmpty(role.ID)
	assert.Equal(roleFormItem.Code, role.Code)
	assert.Equal(roleFormItem.Name, role.Name)
	assert.Equal(roleFormItem.Description, role.Description)
	assert.Equal(roleFormItem.Sequence, role.Sequence)
	assert.Equal(roleFormItem.Status, role.Status)
	assert.Equal(len(roleFormItem.Menus), len(role.Menus))

	// 第三步:创建用户
	// 定义用户表单数据
	userFormItem := schema.UserForm{
		Username: "test",                              // 用户名
		Name:     "Test",                              // 显示名称
		Password: hash.MD5String("test"),              // 密码(MD5加密)
		Phone:    "0720",                              // 电话号码
		Email:    "test@gmail.com",                    // 电子邮箱
		Remark:   "test user",                         // 备注信息
		Status:   schema.UserStatusActivated,          // 用户状态:已激活
		Roles:    schema.UserRoles{{RoleID: role.ID}}, // 用户关联的角色
	}

	// 发送创建用户的POST请求并验证响应
	var user schema.User
	e.POST(baseAPI + "/users").WithJSON(userFormItem).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &user})
	// 验证用户创建结果
	assert.NotEmpty(user.ID)
	assert.Equal(userFormItem.Username, user.Username)
	assert.Equal(userFormItem.Name, user.Name)
	assert.Equal(userFormItem.Phone, user.Phone)
	assert.Equal(userFormItem.Email, user.Email)
	assert.Equal(userFormItem.Remark, user.Remark)
	assert.Equal(userFormItem.Status, user.Status)
	assert.Equal(len(userFormItem.Roles), len(user.Roles))

	// 第四步:查询用户列表
	var users schema.Users
	e.GET(baseAPI+"/users").WithQuery("username", userFormItem.Username).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &users})
	assert.GreaterOrEqual(len(users), 1) // 验证查询结果至少包含一个用户

	// 第五步:更新用户信息
	newName := "Test 1"
	newStatus := schema.UserStatusFreezed // 更新用户状态为冻结
	user.Name = newName
	user.Status = newStatus
	e.PUT(baseAPI + "/users/" + user.ID).WithJSON(user).Expect().Status(http.StatusOK)

	var getUser schema.User
	e.GET(baseAPI + "/users/" + user.ID).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &getUser})
	assert.Equal(newName, getUser.Name)
	assert.Equal(newStatus, getUser.Status)

	// 第六步:清理测试数据
	// 删除用户并验证
	e.DELETE(baseAPI + "/users/" + user.ID).Expect().Status(http.StatusOK)
	e.GET(baseAPI + "/users/" + user.ID).Expect().Status(http.StatusNotFound)

	// 删除角色并验证
	e.DELETE(baseAPI + "/roles/" + role.ID).Expect().Status(http.StatusOK)
	e.GET(baseAPI + "/roles/" + role.ID).Expect().Status(http.StatusNotFound)

	// 删除菜单并验证
	e.DELETE(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusOK)
	e.GET(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusNotFound)
}
