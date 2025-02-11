// package bootstrap 用于应用程序的初始化和引导
package bootstrap

import (
	"context"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/pkg/gormx"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/spf13/cast"
)

// initLoggerHook 初始化日志钩子
// 参数:
//   - ctx: 上下文对象
//   - cfg: 日志钩子配置对象
//
// 返回:
//   - *logging.Hook: 日志钩子实例
//   - error: 错误信息
func initLoggerHook(_ context.Context, cfg *logging.HookConfig) (*logging.Hook, error) {
	// 初始化额外信息映射
	extra := cfg.Extra
	if extra == nil {
		extra = make(map[string]string)
	}
	// 添加应用名称到额外信息中
	extra["appname"] = config.C.General.AppName

	// 根据配置的钩子类型进行不同的初始化
	switch cfg.Type {
	case "gorm": // 如果是 gorm 类型的钩子
		// 创建新的 GORM 数据库连接
		db, err := gormx.New(gormx.Config{
			Debug:        cast.ToBool(cfg.Options["Debug"]),       // 是否开启调试模式
			DBType:       cast.ToString(cfg.Options["DBType"]),    // 数据库类型
			DSN:          cast.ToString(cfg.Options["DSN"]),       // 数据库连接字符串
			MaxLifetime:  cast.ToInt(cfg.Options["MaxLifetime"]),  // 连接最大生命周期
			MaxIdleTime:  cast.ToInt(cfg.Options["MaxIdleTime"]),  // 空闲连接最大生命周期
			MaxOpenConns: cast.ToInt(cfg.Options["MaxOpenConns"]), // 最大打开连接数
			MaxIdleConns: cast.ToInt(cfg.Options["MaxIdleConns"]), // 最大空闲连接数
			TablePrefix:  config.C.Storage.DB.TablePrefix,         // 数据表前缀
		})
		if err != nil {
			return nil, err
		}

		// 创建并返回新的日志钩子
		hook := logging.NewHook(logging.NewGormHook(db),
			logging.SetHookExtra(cfg.Extra),          // 设置额外信息
			logging.SetHookMaxJobs(cfg.MaxBuffer),    // 设置最大缓冲作业数
			logging.SetHookMaxWorkers(cfg.MaxThread), // 设置最大工作线程数
		)
		return hook, nil
	default:
		return nil, nil // 如果是未知的钩子类型，返回空值
	}
}
