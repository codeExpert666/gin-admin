// Package toml 提供了对 TOML 格式数据的编码和解码支持
// TOML (Tom's Obvious, Minimal Language) 是一种配置文件格式
package toml

import (
	"bytes"

	// 导入第三方 TOML 解析库
	"github.com/BurntSushi/toml"
)

// 直接导出 github.com/BurntSushi/toml 包中的核心函数，方便使用
var (
	// Unmarshal 将 TOML 格式的数据解析为 Go 结构体
	Unmarshal = toml.Unmarshal
	// DecodeFile 从文件中读取并解析 TOML 数据
	DecodeFile = toml.DecodeFile
	// Decode 从 Reader 中解析 TOML 数据
	Decode = toml.Decode
)

// Value 类型别名，用于存储 TOML 原始数据
type Value = toml.Primitive

// Marshal 将 Go 数据结构编码为 TOML 格式的字节切片
// 参数 v 可以是结构体、map 等数据类型
// 返回编码后的字节切片和可能的错误
func Marshal(v interface{}) ([]byte, error) {
	// 创建一个新的字节缓冲区
	buf := new(bytes.Buffer)
	// 创建 TOML 编码器并将数据编码写入缓冲区
	err := toml.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalToString 将 Go 数据结构编码为 TOML 格式的字符串
// 是 Marshal 函数的包装，提供更便捷的字符串输出
// 参数 v 可以是结构体、map 等数据类型
// 返回编码后的字符串和可能的错误
func MarshalToString(v interface{}) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
