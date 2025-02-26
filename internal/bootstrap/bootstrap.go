// Package bootstrap 提供了应用程序的引导启动功能
package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // 导入pprof用于性能分析
	"os"
	"strings"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	_ "github.com/LyricTian/gin-admin/v10/internal/swagger" // 导入swagger文档
	"github.com/LyricTian/gin-admin/v10/internal/utility/prom"
	"github.com/LyricTian/gin-admin/v10/internal/wirex"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"go.uber.org/zap"
)

// RunConfig 定义了运行命令所需的配置结构
type RunConfig struct {
	WorkDir   string // 工作目录路径
	Configs   string // 配置文件目录或文件路径（多个配置用逗号分隔）
	StaticDir string // 静态文件目录路径
}

// Run 函数用于初始化和启动服务
// 主要功能包括：
// 1. 加载配置文件
// 2. 初始化日志系统
// 3. 启动pprof调试服务（如果配置）
// 4. 构建依赖注入器
// 5. 初始化Prometheus指标
// 6. 启动HTTP服务
// 7. 处理优雅退出和资源清理
func Run(ctx context.Context, runCfg RunConfig) error {
	// 确保在函数退出时同步日志缓冲区
	defer func() {
		if err := zap.L().Sync(); err != nil {
			fmt.Printf("failed to sync zap logger: %s \n", err.Error())
		}
	}()

	// 加载应用配置
	workDir := runCfg.WorkDir
	staticDir := runCfg.StaticDir
	// 从指定路径加载配置文件，支持多个配置文件（以逗号分隔）
	config.MustLoad(workDir, strings.Split(runCfg.Configs, ",")...)
	// 设置全局工作目录和静态文件目录
	config.C.General.WorkDir = workDir
	config.C.Middleware.Static.Dir = staticDir
	// 打印当前配置信息
	config.C.Print()
	// 执行 Redis 配置复用（验证码、限流器、认证服务）
	config.C.PreLoad()

	// 初始化日志系统
	// 根据配置初始化日志器，并返回清理函数
	cleanLoggerFn, err := logging.InitWithConfig(ctx, &config.C.Logger, initLoggerHook)
	if err != nil {
		return err
	}
	// 为上下文添加主标签，用于日志追踪
	ctx = logging.NewTag(ctx, logging.TagKeyMain)

	logging.Context(ctx).Info("starting service ...",
		zap.String("version", config.C.General.Version),
		zap.Int("pid", os.Getpid()),
		zap.String("workdir", workDir),
		zap.String("config", runCfg.Configs),
		zap.String("static", staticDir),
	)

	// 启动pprof性能分析服务器
	// 如果配置了pprof地址，则在后台启动pprof服务
	if addr := config.C.General.PprofAddr; addr != "" {
		logging.Context(ctx).Info("pprof server is listening on " + addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				logging.Context(ctx).Error("failed to listen pprof server", zap.Error(err))
			}
		}()
	}

	// 构建依赖注入器
	// 使用wire框架构建依赖注入器，返回注入器实例和清理函数
	injector, cleanInjectorFn, err := wirex.BuildInjector(ctx)
	if err != nil {
		return err
	}

	// 初始化注入器
	if err := injector.M.Init(ctx); err != nil {
		return err
	}

	// 初始化全局Prometheus监控指标
	prom.Init()

	// 启动应用服务并处理优雅退出
	return util.Run(ctx, func(ctx context.Context) (func(), error) {
		// 启动HTTP服务器
		cleanHTTPServerFn, err := startHTTPServer(ctx, injector)
		if err != nil {
			return cleanInjectorFn, err
		}

		// 返回清理函数，用于优雅退出时清理资源
		return func() {
			// 释放注入器资源
			if err := injector.M.Release(ctx); err != nil {
				logging.Context(ctx).Error("failed to release injector", zap.Error(err))
			}

			// 按照依赖关系的反序清理资源
			if cleanHTTPServerFn != nil {
				cleanHTTPServerFn()
			}
			if cleanInjectorFn != nil {
				cleanInjectorFn()
			}
			if cleanLoggerFn != nil {
				cleanLoggerFn()
			}
		}, nil
	})
}
