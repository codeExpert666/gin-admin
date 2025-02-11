// package test 定义测试包
package test

import (
	"net/http"
	"testing"

	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/stretchr/testify/assert"
)

// TestMenu 测试菜单管理相关的功能
// 包括:创建菜单、查询菜单列表、更新菜单信息、删除菜单等操作
func TestMenu(t *testing.T) {
	// 初始化测试用的HTTP客户端
	e := tester(t)

	// 创建一个菜单表单数据
	menuFormItem := schema.MenuForm{
		Code:        "menu",                   // 菜单代码
		Name:        "Menu management",        // 菜单名称
		Description: "Menu management",        // 菜单描述
		Sequence:    9,                        // 排序序号
		Type:        "page",                   // 菜单类型:页面
		Path:        "/system/menu",           // 菜单路径
		Properties:  `{"icon":"menu"}`,        // 菜单属性,这里设置了图标
		Status:      schema.MenuStatusEnabled, // 菜单状态:启用
	}

	// 声明一个用于存储创建后的菜单信息的变量
	var menu schema.Menu
	// 发送POST请求创建菜单,并验证返回状态码为200
	e.POST(baseAPI + "/menus").WithJSON(menuFormItem).
		Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &menu})

	// 创建断言对象
	assert := assert.New(t)
	// 验证创建的菜单信息是否符合预期
	assert.NotEmpty(menu.ID)                                 // 确保生成了菜单ID
	assert.Equal(menuFormItem.Code, menu.Code)               // 验证菜单代码
	assert.Equal(menuFormItem.Name, menu.Name)               // 验证菜单名称
	assert.Equal(menuFormItem.Description, menu.Description) // 验证菜单描述
	assert.Equal(menuFormItem.Sequence, menu.Sequence)       // 验证排序序号
	assert.Equal(menuFormItem.Type, menu.Type)               // 验证菜单类型
	assert.Equal(menuFormItem.Path, menu.Path)               // 验证菜单路径
	assert.Equal(menuFormItem.Properties, menu.Properties)   // 验证菜单属性
	assert.Equal(menuFormItem.Status, menu.Status)           // 验证菜单状态

	// 获取菜单列表并验证
	var menus schema.Menus
	e.GET(baseAPI + "/menus").Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &menus})
	assert.GreaterOrEqual(len(menus), 1) // 确保至少有一个菜单项

	// 准备更新菜单信息
	newName := "Menu management 1"         // 新的菜单名称
	newStatus := schema.MenuStatusDisabled // 新的菜单状态:禁用
	menu.Name = newName
	menu.Status = newStatus
	// 发送PUT请求更新菜单信息
	e.PUT(baseAPI + "/menus/" + menu.ID).WithJSON(menu).Expect().Status(http.StatusOK)

	// 获取更新后的菜单信息并验证
	var getMenu schema.Menu
	e.GET(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusOK).JSON().Decode(&util.ResponseResult{Data: &getMenu})
	assert.Equal(newName, getMenu.Name)     // 验证菜单名称已更新
	assert.Equal(newStatus, getMenu.Status) // 验证菜单状态已更新

	// 删除菜单并验证
	e.DELETE(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusOK)
	// 验证删除后无法获取该菜单信息
	e.GET(baseAPI + "/menus/" + menu.ID).Expect().Status(http.StatusNotFound)
}
