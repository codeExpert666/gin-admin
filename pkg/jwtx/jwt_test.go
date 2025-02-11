// Package jwtx JWT认证包的测试文件
package jwtx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert" // 使用 testify 包来进行断言测试
)

// TestAuth 测试 JWT 认证的完整流程
// 包括：
// 1. 生成 token
// 2. 解析 token
// 3. 销毁 token
// 4. 验证已销毁的 token
// 5. 释放资源
func TestAuth(t *testing.T) {
	// 创建一个内存缓存，用于存储已失效的 token
	// CleanupInterval 设置为 1 秒，表示每秒清理一次过期的缓存项
	cache := NewMemoryCache(MemoryConfig{CleanupInterval: time.Second})

	// 使用内存缓存创建一个 token 存储器
	store := NewStoreWithCache(cache)

	// 创建一个空的上下文
	ctx := context.Background()

	// 创建一个新的 JWT 认证器，使用默认配置
	jwtAuth := New(store)

	// 模拟用户ID
	userID := "test"

	// 测试生成 token
	token, err := jwtAuth.GenerateToken(ctx, userID)
	// 断言生成 token 时没有错误发生
	assert.Nil(t, err)
	// 断言生成的 token 不为空
	assert.NotNil(t, token)

	// 测试解析 token
	id, err := jwtAuth.ParseSubject(ctx, token.GetAccessToken())
	// 断言解析 token 时没有错误发生
	assert.Nil(t, err)
	// 断言解析出的用户ID与原始用户ID相同
	assert.Equal(t, userID, id)

	// 测试销毁 token
	err = jwtAuth.DestroyToken(ctx, token.GetAccessToken())
	// 断言销毁 token 时没有错误发生
	assert.Nil(t, err)

	// 测试解析已销毁的 token
	id, err = jwtAuth.ParseSubject(ctx, token.GetAccessToken())
	// 断言解析已销毁的 token 会返回错误
	assert.NotNil(t, err)
	// 断言返回的错误信息为 "Invalid token"
	assert.EqualError(t, err, ErrInvalidToken.Error())
	// 断言返回的用户ID为空
	assert.Empty(t, id)

	// 测试释放资源
	err = jwtAuth.Release(ctx)
	// 断言释放资源时没有错误发生
	assert.Nil(t, err)
}
