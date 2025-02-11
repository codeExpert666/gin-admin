package cachex

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBadgerCache 测试基于 Badger 实现的缓存功能
// 包括测试以下操作：
// 1. 设置缓存值 (Set)
// 2. 获取缓存值 (Get)
// 3. 删除缓存值 (Delete)
// 4. 遍历缓存键值对 (Iterator)
func TestBadgerCache(t *testing.T) {
	// 创建测试断言对象，用于进行测试验证
	assert := assert.New(t)

	// 初始化 BadgerCache 实例，配置数据存储路径
	cache := NewBadgerCache(BadgerConfig{
		Path: "./tmp/badger",
	})

	// 创建上下文对象，用于控制操作的生命周期
	ctx := context.Background()

	// 测试设置缓存值
	// 参数说明：ctx(上下文), "tt"(命名空间), "foo"(键), "bar"(值)
	err := cache.Set(ctx, "tt", "foo", "bar")
	assert.Nil(err) // 确保设置操作没有错误

	// 测试获取缓存值
	// Get 方法返回三个值：值、是否存在、错误信息
	val, exists, err := cache.Get(ctx, "tt", "foo")
	assert.Nil(err)          // 确保获取操作没有错误
	assert.True(exists)      // 确保键值对存在
	assert.Equal("bar", val) // 确保获取的值正确

	// 测试删除缓存值
	err = cache.Delete(ctx, "tt", "foo")
	assert.Nil(err) // 确保删除操作没有错误

	// 验证删除后的状态
	val, exists, err = cache.Get(ctx, "tt", "foo")
	assert.Nil(err)       // 确保获取操作没有错误
	assert.False(exists)  // 确保键值对已被删除
	assert.Equal("", val) // 确保值为空

	// 测试批量设置和遍历功能
	tmap := make(map[string]bool) // 创建用于验证的 map
	// 循环设置 10 个测试数据
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("foo%d", i)
		// 在 "tt" 命名空间下设置数据
		err = cache.Set(ctx, "tt", key, "bar")
		assert.Nil(err)
		tmap[key] = true

		// 在 "ff" 命名空间下设置数据
		err = cache.Set(ctx, "ff", key, "bar")
		assert.Nil(err)
	}

	// 测试遍历功能
	// Iterator 方法接收一个回调函数，用于处理每个键值对
	err = cache.Iterator(ctx, "tt", func(ctx context.Context, key, value string) bool {
		assert.True(tmap[key])     // 验证键是否在预期集合中
		assert.Equal("bar", value) // 验证值是否正确
		t.Log(key, value)          // 记录遍历的键值对
		return true                // 返回 true 继续遍历
	})
	assert.Nil(err) // 确保遍历操作没有错误

	// 关闭缓存实例
	err = cache.Close(ctx)
	assert.Nil(err) // 确保关闭操作没有错误
}
