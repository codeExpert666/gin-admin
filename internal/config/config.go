package config

import (
	"fmt"

	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
)

// Config 应用程序的主配置结构
// 包含了整个应用的所有配置项，包括日志、通用配置、存储、中间件等
type Config struct {
	Logger     logging.LoggerConfig // 日志配置
	General    General              // 通用配置
	Storage    Storage              // 存储配置（缓存和数据库）
	Middleware Middleware           // 中间件配置
	Util       Util                 // 工具配置（验证码、监控等）
	Dictionary Dictionary           // 字典配置
}

// General 通用配置结构
// 包含应用的基本信息和HTTP服务器配置
type General struct {
	AppName            string `default:"ginadmin"` // 应用名称
	Version            string `default:"v10.1.0"`  // 应用版本
	Debug              bool   // 是否开启调试模式
	PprofAddr          string // pprof调试地址
	DisableSwagger     bool   // 是否禁用Swagger文档
	DisablePrintConfig bool   // 是否禁用配置打印
	DefaultLoginPwd    string `default:"6351623c8cef86fefabfa7da046fc619"` // 默认登录密码(MD5加密后的值)
	WorkDir            string // 工作目录（从命令行参数获取）
	MenuFile           string // 菜单配置文件（JSON/YAML格式）
	DenyOperateMenu    bool   // 是否禁止操作菜单
	HTTP               struct {
		Addr            string `default:":8040"` // HTTP服务监听地址
		ShutdownTimeout int    `default:"10"`    // 优雅关闭超时时间（秒）
		ReadTimeout     int    `default:"60"`    // 读取超时时间（秒）
		WriteTimeout    int    `default:"60"`    // 写入超时时间（秒）
		IdleTimeout     int    `default:"10"`    // 空闲连接超时时间（秒）
		CertFile        string // SSL证书文件
		KeyFile         string // SSL密钥文件
	}
	Root struct {
		ID       string `default:"root"`  // 超级管理员ID
		Username string `default:"admin"` // 超级管理员用户名
		Password string // 超级管理员密码
		Name     string `default:"Admin"` // 超级管理员显示名称
	}
}

// Storage 存储配置结构
// 包含缓存和数据库的配置信息
type Storage struct {
	Cache struct {
		Type      string `default:"memory"` // 缓存类型（memory/badger/redis）
		Delimiter string `default:":"`      // 缓存键分隔符
		Memory    struct {
			CleanupInterval int `default:"60"` // 内存缓存清理间隔（秒）
		}
		Badger struct {
			Path string `default:"data/cache"` // Badger数据库存储路径
		}
		Redis struct {
			Addr     string // Redis服务器地址
			Username string // Redis用户名
			Password string // Redis密码
			DB       int    // Redis数据库索引
		}
	}
	DB struct {
		Debug        bool   // 是否开启调试模式
		Type         string `default:"sqlite3"`          // 数据库类型（sqlite3/mysql/postgres）
		DSN          string `default:"data/ginadmin.db"` // 数据库连接字符串
		MaxLifetime  int    `default:"86400"`            // 连接最大生命周期（秒）
		MaxIdleTime  int    `default:"3600"`             // 空闲连接超时时间（秒）
		MaxOpenConns int    `default:"100"`              // 最大打开连接数
		MaxIdleConns int    `default:"50"`               // 最大空闲连接数
		TablePrefix  string `default:""`                 // 数据库表前缀
		AutoMigrate  bool   // 是否自动迁移数据库表
		PrepareStmt  bool   // 是否启用预编译语句
		Resolver     []struct {
			DBType   string   // 数据库类型
			Sources  []string // 主库DSN列表
			Replicas []string // 从库DSN列表
			Tables   []string // 相关的表名列表
		}
	}
}

// Util 工具配置结构
// 包含验证码和Prometheus监控的配置
type Util struct {
	Captcha struct {
		Length    int    `default:"4"`      // 验证码长度
		Width     int    `default:"400"`    // 验证码图片宽度
		Height    int    `default:"160"`    // 验证码图片高度
		CacheType string `default:"memory"` // 验证码缓存类型（memory/redis）
		Redis     struct {
			Addr      string // Redis服务器地址
			Username  string // Redis用户名
			Password  string // Redis密码
			DB        int    // Redis数据库索引
			KeyPrefix string `default:"captcha:"` // 验证码键前缀
		}
	}
	Prometheus struct {
		Enable         bool     // 是否启用Prometheus监控
		Port           int      `default:"9100"`  // Prometheus监控端口
		BasicUsername  string   `default:"admin"` // 基本认证用户名
		BasicPassword  string   `default:"admin"` // 基本认证密码
		LogApis        []string // 需要记录的API路径
		LogMethods     []string // 需要记录的HTTP方法
		DefaultCollect bool     // 是否启用默认的监控指标
	}
}

// Dictionary 字典配置结构
type Dictionary struct {
	UserCacheExp int `default:"4"` // 用户缓存过期时间（小时）
}

// IsDebug 判断是否为调试模式
func (c *Config) IsDebug() bool {
	return c.General.Debug
}

// String 将配置转换为JSON字符串
func (c *Config) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic("Failed to marshal config: " + err.Error())
	}
	return string(b)
}

// PreLoad Redis配置自动复用
// 如果设置了主Redis配置，会自动复制到其他需要Redis的组件中（验证码、限流器、认证服务）
func (c *Config) PreLoad() {
	if addr := c.Storage.Cache.Redis.Addr; addr != "" {
		username := c.Storage.Cache.Redis.Username
		password := c.Storage.Cache.Redis.Password
		// Redis配置复制到验证码服务
		if c.Util.Captcha.CacheType == "redis" &&
			c.Util.Captcha.Redis.Addr == "" {
			c.Util.Captcha.Redis.Addr = addr
			c.Util.Captcha.Redis.Username = username
			c.Util.Captcha.Redis.Password = password
		}
		// Redis配置复制到限流器
		if c.Middleware.RateLimiter.Store.Type == "redis" &&
			c.Middleware.RateLimiter.Store.Redis.Addr == "" {
			c.Middleware.RateLimiter.Store.Redis.Addr = addr
			c.Middleware.RateLimiter.Store.Redis.Username = username
			c.Middleware.RateLimiter.Store.Redis.Password = password
		}
		// Redis配置复制到认证服务
		if c.Middleware.Auth.Store.Type == "redis" &&
			c.Middleware.Auth.Store.Redis.Addr == "" {
			c.Middleware.Auth.Store.Redis.Addr = addr
			c.Middleware.Auth.Store.Redis.Username = username
			c.Middleware.Auth.Store.Redis.Password = password
		}
	}
}

// Print 打印配置信息
// 如果DisablePrintConfig为true则不打印
func (c *Config) Print() {
	if c.General.DisablePrintConfig {
		return
	}
	fmt.Println("// ----------------------- Load configurations start ------------------------")
	fmt.Println(c.String())
	fmt.Println("// ----------------------- Load configurations end --------------------------")
}

// FormatTableName 格式化表名
// 返回带有表前缀的完整表名
func (c *Config) FormatTableName(name string) string {
	return c.Storage.DB.TablePrefix + name
}
