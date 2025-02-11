// Package jwtx 提供了 JWT (JSON Web Token) 相关的功能实现
package jwtx

import (
	// 导入 jsoniter 包来处理 JSON 序列化和反序列化
	// jsoniter 是一个高性能的 JSON 解析库
	jsoniter "github.com/json-iterator/go"
)

// TokenInfo 接口定义了 Token 信息的基本操作方法
// 这个接口可以被不同的 Token 实现结构体所实现
type TokenInfo interface {
	// GetAccessToken 返回访问令牌字符串
	GetAccessToken() string
	// GetTokenType 返回令牌类型（如 "Bearer"）
	GetTokenType() string
	// GetExpiresAt 返回令牌的过期时间戳（Unix 时间戳格式）
	GetExpiresAt() int64
	// EncodeToJSON 将 Token 信息编码为 JSON 字节数组
	EncodeToJSON() ([]byte, error)
}

// tokenInfo 结构体实现了 TokenInfo 接口
// 用于存储 Token 的具体信息
type tokenInfo struct {
	AccessToken string `json:"access_token"` // 访问令牌字符串
	TokenType   string `json:"token_type"`   // 令牌类型，通常为 "Bearer"
	ExpiresAt   int64  `json:"expires_at"`   // 令牌过期时间（Unix 时间戳）
}

// GetAccessToken 实现 TokenInfo 接口，返回访问令牌
func (t *tokenInfo) GetAccessToken() string {
	return t.AccessToken
}

// GetTokenType 实现 TokenInfo 接口，返回令牌类型
func (t *tokenInfo) GetTokenType() string {
	return t.TokenType
}

// GetExpiresAt 实现 TokenInfo 接口，返回过期时间
func (t *tokenInfo) GetExpiresAt() int64 {
	return t.ExpiresAt
}

// EncodeToJSON 实现 TokenInfo 接口，将 token 信息序列化为 JSON 格式
// 返回值：
// - []byte: JSON 格式的字节数组
// - error: 序列化过程中可能发生的错误
func (t *tokenInfo) EncodeToJSON() ([]byte, error) {
	return jsoniter.Marshal(t)
}
