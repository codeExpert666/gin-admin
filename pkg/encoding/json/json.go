// Package json 提供了JSON序列化和反序列化的功能封装
// 本包使用 json-iterator/go 库来提供更高性能的JSON处理能力
package json

import (
	"fmt"

	// 导入 json-iterator 库，这是一个高性能的 JSON 处理库
	// 相比标准库具有更好的性能表现
	jsoniter "github.com/json-iterator/go"
)

var (
	// 创建一个与标准库兼容的 JSON 配置实例
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	// Marshal 将 Go 数据结构序列化为 JSON 字节切片
	Marshal = json.Marshal

	// Unmarshal 将 JSON 字节切片反序列化为 Go 数据结构
	Unmarshal = json.Unmarshal

	// MarshalIndent 将 Go 数据结构序列化为格式化的 JSON 字节切片
	// 主要用于优化输出的可读性
	MarshalIndent = json.MarshalIndent

	// NewDecoder 创建一个新的 JSON 解码器
	// 用于从 io.Reader 中读取和解码 JSON 数据
	NewDecoder = json.NewDecoder

	// NewEncoder 创建一个新的 JSON 编码器
	// 用于将 JSON 数据写入 io.Writer
	NewEncoder = json.NewEncoder
)

// MarshalToString 将任意类型的数据转换为 JSON 字符串
// 参数 v：要转换的数据（interface{} 表示可以接受任意类型）
// 返回值：转换后的 JSON 字符串
func MarshalToString(v interface{}) string {
	// 使用 jsoniter 的 MarshalToString 方法进行转换
	s, err := jsoniter.MarshalToString(v)
	if err != nil {
		// 如果转换过程中发生错误，打印错误信息并返回空字符串
		fmt.Println("Failed to marshal json string: " + err.Error())
		return ""
	}
	return s
}
