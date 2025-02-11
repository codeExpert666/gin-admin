package cachex

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedisCache 测试Redis缓存的基本功能
// 包括：设置值、获取值、删除值和遍历操作
func TestRedisCache(t *testing.T) {
	// 创建一个断言对象，用于测试结果验证
	assert := assert.New(t)

	// 创建一个Redis缓存实例
	// 配置连接到本地Redis服务器，使用数据库1
	cache := NewRedisCache(RedisConfig{
		Addr: "localhost:6379", // Redis服务器地址
		DB:   1,                // 使用的数据库编号
	})

	// 创建一个背景上下文
	ctx := context.Background()

	// 测试1：设置和获取单个键值对
	// 在命名空间"tt"下设置键"foo"的值为"bar"
	err := cache.Set(ctx, "tt", "foo", "bar")
	assert.Nil(err) // 确保设置操作没有错误

	// 获取刚才设置的值
	val, exists, err := cache.Get(ctx, "tt", "foo")
	assert.Nil(err)          // 确保获取操作没有错误
	assert.True(exists)      // 确保键值对存在
	assert.Equal("bar", val) // 确保获取的值正确

	// 测试2：删除键值对
	err = cache.Delete(ctx, "tt", "foo")
	assert.Nil(err) // 确保删除操作没有错误

	// 验证删除后无法获取该键值对
	val, exists, err = cache.Get(ctx, "tt", "foo")
	assert.Nil(err)       // 确保获取操作没有错误
	assert.False(exists)  // 确保键值对不存在
	assert.Equal("", val) // 确保返回空值

	// 测试3：批量操作和遍历
	// 创建一个map用于验证遍历结果
	tmap := make(map[string]bool)

	// 在两个不同的命名空间（tt和ff）中设置10个键值对
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("foo%d", i)
		// 在tt命名空间设置值
		err = cache.Set(ctx, "tt", key, "bar")
		assert.Nil(err)
		tmap[key] = true

		// 在ff命名空间设置值
		err = cache.Set(ctx, "ff", key, "bar")
		assert.Nil(err)
	}

	// 测试4：遍历指定命名空间(tt)中的所有键值对
	err = cache.Iterator(ctx, "tt", func(ctx context.Context, key, value string) bool {
		assert.True(tmap[key])     // 验证键是否在预期map中
		assert.Equal("bar", value) // 验证值是否正确
		return true                // 返回true继续遍历
	})
	assert.Nil(err) // 确保遍历操作没有错误

	// 测试5：关闭缓存连接
	err = cache.Close(ctx)
	assert.Nil(err) // 确保关闭操作没有错误
}
