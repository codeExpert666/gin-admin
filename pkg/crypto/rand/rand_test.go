// Package rand 的测试文件，用于验证随机字符串生成功能的正确性
package rand

import (
	"strconv"
	"testing"
)

// TestRandom 测试Random函数的功能
// 测试用例：生成一个长度为6的纯数字随机字符串
// 验证内容：
// 1. 确保不会返回错误
// 2. 验证生成的字符串长度是否正确
// 3. 验证生成的每个字符是否都是有效的数字（0-9）
func TestRandom(t *testing.T) {
	// 生成长度为6的纯数字随机字符串
	digits, err := Random(6, Ldigit)
	if err != nil {
		t.Error("生成随机字符串时发生错误:", err.Error())
		return
	} else if len(digits) != 6 {
		t.Error("生成的字符串长度不符合预期:", digits)
		return
	}

	// 遍历生成的字符串中的每个字符
	for _, b := range digits {
		// 将字符转换为数字
		d, err := strconv.Atoi(string(b))
		if err != nil {
			t.Error("字符转换为数字时发生错误:", err.Error())
			return
		} else if d > 10 || d < 0 {
			// 验证数字是否在有效范围内（0-9）
			t.Error("生成的数字超出有效范围:", d)
		}
	}
}
