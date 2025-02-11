// Package util 提供了一些通用的工具函数
package util

import (
	"strings" // 导入strings包用于字符串处理
	"testing" // 导入testing包用于编写测试用例
)

// TestNewXID 测试生成XID的功能
// XID是一个全局唯一的ID生成器,生成的ID格式为20字节的字符串
// 参数t *testing.T 是Go测试框架提供的测试对象,用于记录测试结果
func TestNewXID(t *testing.T) {
	// 调用NewXID()生成一个新的XID,并转换为大写格式后打印到测试日志
	t.Logf("xid: %s", strings.ToUpper(NewXID()))
}

// TestMustNewUUID 测试生成UUID的功能
// UUID是一个符合RFC 4122标准的通用唯一标识符
// 参数t *testing.T 是Go测试框架提供的测试对象,用于记录测试结果
func TestMustNewUUID(t *testing.T) {
	// 调用MustNewUUID()生成一个新的UUID,并转换为大写格式后打印到测试日志
	// MustNewUUID()函数在生成UUID失败时会直接panic,适用于确保必须生成成功的场景
	t.Logf("uuid: %s", strings.ToUpper(MustNewUUID()))
}
