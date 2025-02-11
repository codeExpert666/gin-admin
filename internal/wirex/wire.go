// 这是一个特殊的构建标签，表示这个文件只在执行 wire 命令生成代码时使用
//go:build wireinject
// +build wireinject

package wirex

// 上面的构建标签确保这个存根代码不会包含在最终构建中

import (
	"context"

	// google/wire 是 Google 开发的依赖注入框架
	"github.com/google/wire"

	// 导入项目内部的模块
	"github.com/LyricTian/gin-admin/v10/internal/mods"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
)

// BuildInjector 是一个依赖注入的构建函数
// 它使用 wire 框架来自动生成依赖注入的代码
// 参数:
//   - ctx: 上下文对象，用于传递超时、取消信号等
//
// 返回值:
//   - *Injector: 注入器实例
//   - func(): 清理函数，用于资源释放
//   - error: 可能发生的错误
func BuildInjector(ctx context.Context) (*Injector, func(), error) {
	// wire.Build 用于声明所有需要注入的依赖
	wire.Build(
		// 初始化缓存服务
		InitCacher,
		// 初始化数据库连接
		InitDB,
		// 初始化认证服务
		InitAuth,
		// 使用 wire.NewSet 创建新的提供者集合
		// wire.Struct 用于创建结构体实例，"*" 表示注入所有字段
		wire.NewSet(wire.Struct(new(util.Trans), "*")),
		// 创建 Injector 结构体实例
		wire.NewSet(wire.Struct(new(Injector), "*")),
		// 注入 mods 包中定义的所有依赖
		mods.Set,
	) // 依赖注入声明结束

	// 这个返回语句实际上不会被执行
	// wire 工具会根据上面的 Build 声明生成真实的实现代码
	return new(Injector), nil, nil
}
