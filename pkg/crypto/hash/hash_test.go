// Package hash 的测试文件
// 包含了对 hash 包中各个函数的单元测试
package hash

import (
	"testing" // 导入 Go 语言的测试包
)

// TestGeneratePassword 测试密码哈希生成和验证功能
// 测试步骤：
// 1. 使用一个测试密码生成哈希值
// 2. 验证生成的哈希值是否可以正确匹配原始密码
// 3. 同时输出哈希后的密码长度，用于观察 bcrypt 的输出特征
func TestGeneratePassword(t *testing.T) {
	// 定义测试用的原始密码
	origin := "abc-123"

	// 使用 GeneratePassword 生成密码哈希
	hashPwd, err := GeneratePassword(origin)
	if err != nil {
		t.Error("GeneratePassword Failed: ", err.Error())
	}

	// 输出生成的哈希密码和其长度，便于调试
	t.Log("test password: ", hashPwd, ",length: ", len(hashPwd))

	// 验证生成的哈希密码是否能够正确匹配原始密码
	if err := CompareHashAndPassword(hashPwd, origin); err != nil {
		t.Error("Unmatched password: ", err.Error())
	}
}

// TestMD5 测试 MD5 哈希函数的正确性
// 测试方法：
// 1. 使用预先计算好的已知输入和输出进行对比
// 2. 如果生成的哈希值与预期值不匹配，则测试失败
func TestMD5(t *testing.T) {
	// 测试用的原始字符串
	origin := "abc-123"
	// 预期的 MD5 哈希值（使用其他工具预先计算得到）
	hashVal := "6351623c8cef86fefabfa7da046fc619"

	// 使用 MD5String 函数计算哈希值，并与预期值比较
	if v := MD5String(origin); v != hashVal {
		t.Error("Failed to generate MD5 hash: ", v)
	}
}
