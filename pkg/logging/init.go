// Package logging 提供了一个灵活的日志系统实现
package logging

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"     // 用于解析 TOML 配置文件
	"go.uber.org/zap"                  // 高性能的结构化日志库
	"go.uber.org/zap/zapcore"          // zap 的核心组件
	"gopkg.in/natefinch/lumberjack.v2" // 用于日志文件切割和管理
)

// Config 是总配置结构体
type Config struct {
	Logger LoggerConfig // 日志配置部分
}

// LoggerConfig 定义了日志系统的详细配置项
type LoggerConfig struct {
	Debug      bool     // 是否为调试模式（为 true 时，使用 zap 开发配置；为 false，使用 zap 生产配置）
	Level      string   // 日志级别 debug（调试信息）/info（一般信息）/warn（警告信息）/error（错误信息）/dpanic（开发环境会 panic）/panic（会 panic）/fatal（致命错误，记录后程序退出）
	CallerSkip int      // 调用栈跳过层数，帮助确定日志输出时显示的代码位置更准确
	File       struct { // 文件输出配置
		Enable     bool   // 是否启用文件输出
		Path       string // 日志文件路径
		MaxSize    int    // 单个日志文件最大尺寸（MB）
		MaxBackups int    // 保留的旧日志文件数量
	}
	Hooks []*HookConfig // 日志钩子配置，使用指针切片允许动态修改钩子配置
}

// HookConfig 定义了单个日志钩子的配置
type HookConfig struct {
	Enable    bool              // 是否启用该钩子
	Level     string            // 日志级别，该钩子仅处理达到或超过该级别的日志
	Type      string            // 钩子类型，例如 "gorm" 用于数据库日志
	MaxBuffer int               // 缓冲区最大容量，用于批量处理日志
	MaxThread int               // 处理日志的最大线程数
	Options   map[string]string // 钩子的配置选项
	Extra     map[string]string // 额外的自定义配置项
}

// HookHandlerFunc 定义了注册日志钩子处理器的函数类型
// ctx: 上下文信息
// hookCfg: 钩子配置
// 返回值: 钩子实例和可能的错误
type HookHandlerFunc func(ctx context.Context, hookCfg *HookConfig) (*Hook, error)

// LoadConfigFromToml 从 TOML 文件加载日志配置
// filename: TOML 配置文件的路径
// 返回值: 日志配置和可能的错误
func LoadConfigFromToml(filename string) (*LoggerConfig, error) {
	cfg := &Config{}
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}
	return &cfg.Logger, nil
}

// InitWithConfig 使用给定的配置初始化日志系统
// ctx: 上下文信息
// cfg: 日志配置
// hookHandle: 可选的钩子处理函数
// 返回值: 清理函数和可能的错误
func InitWithConfig(ctx context.Context, cfg *LoggerConfig, hookHandle ...HookHandlerFunc) (func(), error) {
	// 根据是否为调试模式选择 zap 的配置
	var zconfig zap.Config
	if cfg.Debug {
		cfg.Level = "debug"
		zconfig = zap.NewDevelopmentConfig() // 开发环境配置：更详细的日志输出
	} else {
		zconfig = zap.NewProductionConfig() // 生产环境配置：注重性能
	}

	// 解析并设置日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zconfig.Level.SetLevel(level)

	var (
		logger   *zap.Logger
		cleanFns []func() // 存储清理函数的切片
	)

	// 配置文件输出
	if cfg.File.Enable {
		filename := cfg.File.Path
		_ = os.MkdirAll(filepath.Dir(filename), 0777) // 确保日志文件目录存在
		// 配置 lumberjack 进行日志文件管理
		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    cfg.File.MaxSize,    // 单个文件最大尺寸（MB）
			MaxBackups: cfg.File.MaxBackups, // 保留的旧文件数量
			Compress:   false,               // 是否压缩旧文件
			LocalTime:  true,                // 使用本地时间
		}

		// 添加文件关闭到清理函数
		cleanFns = append(cleanFns, func() {
			_ = fileWriter.Close()
		})

		// 创建文件输出的 zapcore
		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(zconfig.EncoderConfig),
			zapcore.AddSync(fileWriter),
			zconfig.Level,
		)
		logger = zap.New(zc)
	} else {
		// 如果不启用文件输出，使用标准配置
		ilogger, err := zconfig.Build()
		if err != nil {
			return nil, err
		}
		logger = ilogger
	}

	// 设置调用栈跳过层数
	skip := cfg.CallerSkip
	if skip <= 0 {
		skip = 2
	}

	// 配置日志选项
	logger = logger.WithOptions(
		zap.WithCaller(true),              // 启用调用者信息
		zap.AddStacktrace(zap.ErrorLevel), // 错误级别及以上添加堆栈跟踪
		zap.AddCallerSkip(skip),           // 设置调用栈跳过层数
	)

	// 配置日志钩子
	for _, h := range cfg.Hooks {
		if !h.Enable || len(hookHandle) == 0 {
			continue
		}

		// 创建钩子写入器
		writer, err := hookHandle[0](ctx, h)
		if err != nil {
			return nil, err
		} else if writer == nil {
			continue
		}

		// 添加钩子刷新到清理函数
		cleanFns = append(cleanFns, func() {
			writer.Flush()
		})

		// 配置钩子的日志级别
		hookLevel := zap.NewAtomicLevel()
		if level, err := zapcore.ParseLevel(h.Level); err == nil {
			hookLevel.SetLevel(level)
		} else {
			hookLevel.SetLevel(zap.InfoLevel)
		}

		// 配置钩子的编码器
		hookEncoder := zap.NewProductionEncoderConfig()
		hookEncoder.EncodeTime = zapcore.EpochMillisTimeEncoder
		hookEncoder.EncodeDuration = zapcore.MillisDurationEncoder
		hookCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(hookEncoder),
			zapcore.AddSync(writer),
			hookLevel,
		)

		// 将钩子核心添加到日志器
		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, hookCore)
		}))
	}

	// 替换全局日志器
	zap.ReplaceGlobals(logger)

	// 返回清理函数
	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}
