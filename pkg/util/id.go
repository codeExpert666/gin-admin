// Package util 提供了一些通用的工具函数
package util

import (
	"github.com/google/uuid" // 导入 uuid 包,用于生成 UUID
	"github.com/rs/xid"      // 导入 xid 包,用于生成 XID
)

// NewXID 生成一个新的 XID(globally unique identifier)并返回其字符串表示
// XID 是一个 12 字节的唯一标识符,比 UUID 更短,且包含时间戳信息
// 它由以下部分组成:
// - 4 字节的 UNIX 时间戳
// - 3 字节的机器标识
// - 2 字节的进程 ID
// - 3 字节的计数器
func NewXID() string {
	return xid.New().String()
}

// MustNewUUID 生成一个新的 UUID(Universally Unique Identifier)并返回其字符串表示
// UUID 是一个 16 字节的唯一标识符,广泛用于分布式系统中
// 如果生成 UUID 时发生错误,该函数会触发 panic
// 注意:在不能接受 panic 的场景下,应该使用其他方式处理错误
func MustNewUUID() string {
	v, err := uuid.NewRandom() // 生成一个随机的 UUID
	if err != nil {
		panic(err) // 如果生成失败,直接触发 panic
	}
	return v.String() // 将 UUID 转换为字符串格式并返回
}
