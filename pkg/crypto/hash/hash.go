// Package hash 提供了常用的哈希函数实现
// 包括：MD5、SHA1 和基于 bcrypt 的密码哈希
package hash

import (
	"crypto/md5"    // 导入标准库 MD5 哈希算法
	"crypto/sha1"   // 导入标准库 SHA1 哈希算法
	"fmt"          // 用于格式化输出

	"golang.org/x/crypto/bcrypt" // 导入 bcrypt 密码哈希算法，这是一个专门用于密码加密的算法
)

// MD5 函数接收一个字节切片作为输入，返回其 MD5 哈希值的十六进制字符串表示
// MD5 是一种广泛使用的哈希函数，输出长度为 128 位（16 字节）
// 注意：MD5 不应该用于安全性要求高的场景，因为它已经被证明不够安全
func MD5(b []byte) string {
	h := md5.New()         // 创建一个新的 MD5 哈希对象
	_, _ = h.Write(b)      // 写入数据到哈希对象
	return fmt.Sprintf("%x", h.Sum(nil)) // 计算哈希值并转换为十六进制字符串
}

// MD5String 是 MD5 函数的便捷包装器，直接接收字符串输入
// 内部会将字符串转换为字节切片后调用 MD5 函数
func MD5String(s string) string {
	return MD5([]byte(s))
}

// SHA1 函数接收一个字节切片作为输入，返回其 SHA1 哈希值的十六进制字符串表示
// SHA1 哈希函数输出长度为 160 位（20 字节）
// 注意：和 MD5 一样，SHA1 也不应该用于安全性要求高的场景
func SHA1(b []byte) string {
	h := sha1.New()        // 创建一个新的 SHA1 哈希对象
	_, _ = h.Write(b)      // 写入数据到哈希对象
	return fmt.Sprintf("%x", h.Sum(nil)) // 计算哈希值并转换为十六进制字符串
}

// SHA1String 是 SHA1 函数的便捷包装器，直接接收字符串输入
// 内部会将字符串转换为字节切片后调用 SHA1 函数
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// GeneratePassword 使用 bcrypt 算法对密码进行哈希处理
// bcrypt 是专门为密码哈希设计的算法，具有以下特点：
// 1. 自动加盐：每次生成的哈希值都不同，即使是相同的密码
// 2. 可调整的工作因子：通过 DefaultCost 可以调整哈希的计算强度
// 3. 内置了防止时序攻击的机制
func GeneratePassword(password string) (string, error) {
	// GenerateFromPassword 使用默认的计算强度（10）生成哈希值
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CompareHashAndPassword 使用 bcrypt 比较哈希密码和明文密码是否匹配
// 参数说明：
//   - hashedPassword: 之前通过 GeneratePassword 生成的哈希密码
//   - password: 需要验证的明文密码
// 返回值：
//   - 如果密码匹配，返回 nil
//   - 如果密码不匹配，返回 error
func CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
