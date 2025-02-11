// Package aes 的测试文件
// 本文件包含了对 AES 加密解密功能的单元测试
package aes

import (
	"testing" // Go 标准库测试包

	"github.com/stretchr/testify/assert" // 使用 testify 提供的断言功能，使测试更加简洁和可读
)

// TestAESEncrypt 测试 AES 加密解密功能的完整流程
// 测试内容包括：
// 1. 使用 Base64 编码的加密功能
// 2. 从 Base64 字符串解密的功能
// 3. 验证解密后的数据与原始数据是否一致
func TestAESEncrypt(t *testing.T) {
	// 创建断言对象，用于进行测试断言
	assert := assert.New(t)

	// 准备测试数据
	data := []byte("hello world")

	// 测试加密功能
	// 将数据加密并转换为 Base64 字符串
	bs64, err := EncryptToBase64(data, SecretKey)
	// 断言加密过程没有错误
	assert.Nil(err)
	// 断言加密结果不为空
	assert.NotEmpty(bs64)

	// 输出加密后的 Base64 字符串，方便调试
	t.Log(bs64)

	// 测试解密功能
	// 将 Base64 字符串解密回原始数据
	result, err := DecryptFromBase64(bs64, SecretKey)
	// 断言解密过程没有错误
	assert.Nil(err)
	// 断言解密后的数据与原始数据完全相同
	assert.Equal(data, result)
}
