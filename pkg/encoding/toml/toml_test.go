// Package toml 的测试文件
package toml

import "testing"

// TestTomlDecode 测试 TOML 解码功能
// 主要测试嵌套结构体和自定义 Value 类型的解析
func TestTomlDecode(t *testing.T) {
	// 定义一个用于测试的匿名结构体
	// 该结构体模拟了中间件配置的数据结构
	var config struct {
		// Middlewares 是一个中间件配置数组
		// 每个中间件包含名称和选项
		Middlewares []struct {
			// Name 表示中间件的名称
			// `toml:"name"` 表示 TOML 文件中对应的键名为 "name"
			Name string `toml:"name"`
			// Options 使用 Value(toml.Primitive) 类型存储原始 TOML 数据
			// 这允许后续进行延迟解析
			Options Value `toml:"options"`
		} `toml:"middlewares"`
	}

	// 使用 Decode 函数解析 TOML 格式的字符串
	// md 包含元数据信息，可用于后续的 PrimitiveDecode
	md, err := Decode(`
	middlewares = [
  		{name = "ratelimit", options = {max = 10, period = 10}},
	]
	`, &config)
	// 检查解析是否出现错误
	if err != nil {
		t.Error(err)
		return
	}

	// 定义限流中间件的配置结构
	var rateLimitConfig struct {
		// Max 表示最大请求数
		Max int `toml:"max"`
		// Period 表示时间周期（单位：秒）
		Period int `toml:"period"`
	}

	// 使用 PrimitiveDecode 解析之前保存的原始 TOML 数据
	// 将第一个中间件的 options 解析到 rateLimitConfig 结构体中
	err = md.PrimitiveDecode(config.Middlewares[0].Options, &rateLimitConfig)
	if err != nil {
		t.Error(err)
		return
	}

	// 验证解析后的配置值是否符合预期
	if rateLimitConfig.Max != 10 || rateLimitConfig.Period != 10 {
		t.Errorf("Expected {Max: 10, Period: 10}, got %v", rateLimitConfig)
	}
}
