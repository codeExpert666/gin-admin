// Package util 提供了一些通用的工具函数
package util

import (
	"encoding/binary" // 导入binary包用于处理二进制数据
	"math/rand"       // 导入rand包用于生成随机数
	"strconv"         // 导入strconv包用于字符串和基本数据类型之间的转换
	"strings"         // 导入strings包用于字符串处理
	"time"            // 导入time包用于获取当前时间
)

// RandomizedIPAddr 生成一个随机的IP地址字符串
// 返回值格式为:xxx.xxx.xxx.xxx,其中xxx为0-255之间的随机数
func RandomizedIPAddr() string {
	// 创建一个4字节的字节数组,用于存储IP地址的4个段
	raw := make([]byte, 4)

	// 使用当前时间的纳秒数作为随机数种子,创建一个新的随机数生成器
	// 这样可以确保每次运行生成不同的随机数
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 使用随机数生成器生成一个32位无符号整数,并将其按小端序写入字节数组
	binary.LittleEndian.PutUint32(raw, rd.Uint32())

	// 创建一个字符串切片,用于存储转换后的IP地址各段
	ips := make([]string, len(raw))

	// 遍历字节数组,将每个字节转换为0-255之间的字符串
	for i, b := range raw {
		ips[i] = strconv.FormatInt(int64(b), 10)
	}

	// 使用点号连接所有IP段,返回最终的IP地址字符串
	return strings.Join(ips, ".")
}
