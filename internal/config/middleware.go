// package config 定义了应用程序的配置结构
package config

// Middleware 结构体定义了所有中间件的配置选项
type Middleware struct {
	// Recovery 定义了恢复中间件的配置
	// 用于捕获程序运行时的 panic，并恢复程序的正常运行
	Recovery struct {
		Skip int `default:"3"` // 跳过前 n 个堆栈帧，用于在错误输出中隐藏框架相关的调用信息
	}
	// CORS 定义了跨域资源共享(Cross-Origin Resource Sharing)的配置
	CORS struct {
		Enable                 bool      // 是否启用 CORS
		AllowAllOrigins        bool      // 是否允许所有来源的请求
		AllowOrigins           []string  // 允许的来源域名列表
		AllowMethods           []string  // 允许的 HTTP 方法列表（如 GET, POST, PUT 等）
		AllowHeaders           []string  // 允许的 HTTP 请求头列表
		AllowCredentials       bool      // 是否允许发送认证信息（cookies）
		ExposeHeaders          []string  // 允许浏览器访问的响应头列表
		MaxAge                 int       // 预检请求的缓存时间（秒）
		AllowWildcard          bool      // 是否允许通配符匹配域名
		AllowBrowserExtensions bool      // 是否允许浏览器扩展访问
		AllowWebSockets        bool      // 是否允许 WebSocket 连接
		AllowFiles             bool      // 是否允许文件请求
	}
	// Trace 定义了请求追踪的配置
	Trace struct {
		SkippedPathPrefixes []string  // 不需要追踪的路径前缀列表
		RequestHeaderKey    string `default:"X-Request-Id"` // 请求头中的追踪 ID 键名
		ResponseTraceKey    string `default:"X-Trace-Id"`  // 响应头中的追踪 ID 键名
	}
	// Logger 定义了日志中间件的配置
	Logger struct {
		SkippedPathPrefixes      []string  // 不需要记录日志的路径前缀列表
		MaxOutputRequestBodyLen  int `default:"4096"`  // 请求体日志最大长度（字节）
		MaxOutputResponseBodyLen int `default:"1024"` // 响应体日志最大长度（字节）
	}
	// CopyBody 定义了请求体复制中间件的配置
	CopyBody struct {
		SkippedPathPrefixes []string  // 不需要复制请求体的路径前缀列表
		MaxContentLen       int64 `default:"33554432"` // 最大内容长度（默认 32MB）
	}
	// Auth 定义了认证中间件的配置
	Auth struct {
		Disable             bool      // 是否禁用认证
		SkippedPathPrefixes []string  // 不需要认证的路径前缀列表
		SigningMethod       string `default:"HS512"`    // JWT 签名方法（支持 HS256/HS384/HS512）
		SigningKey          string `default:"XnEsT0S@"` // JWT 签名密钥
		OldSigningKey       string // 旧的签名密钥（用于密钥迁移）
		Expired             int    `default:"86400"` // Token 过期时间（秒）
		Store               struct {
			Type      string `default:"memory"` // 存储类型（支持 memory/badger/redis）
			Delimiter string `default:":"`      // 键名分隔符
			// Memory 存储配置
			Memory    struct {
				CleanupInterval int `default:"60"` // 清理间隔（秒）
			}
			// Badger 存储配置
			Badger struct {
				Path string `default:"data/auth"` // 数据存储路径
			}
			// Redis 存储配置
			Redis struct {
				Addr     string // Redis 服务器地址
				Username string // Redis 用户名
				Password string // Redis 密码
				DB       int    // Redis 数据库索引
			}
		}
	}
	// RateLimiter 定义了限流中间件的配置
	RateLimiter struct {
		Enable              bool      // 是否启用限流
		SkippedPathPrefixes []string  // 不需要限流的路径前缀列表
		Period              int       // 限流周期（秒）
		MaxRequestsPerIP    int       // 每个 IP 在周期内的最大请求次数
		MaxRequestsPerUser  int       // 每个用户在周期内的最大请求次数
		Store               struct {
			Type   string // 存储类型（支持 memory/redis）
			// Memory 存储配置
			Memory struct {
				Expiration      int `default:"3600"` // 过期时间（秒）
				CleanupInterval int `default:"60"`   // 清理间隔（秒）
			}
			// Redis 存储配置
			Redis struct {
				Addr     string // Redis 服务器地址
				Username string // Redis 用户名
				Password string // Redis 密码
				DB       int    // Redis 数据库索引
			}
		}
	}
	// Casbin 定义了访问控制中间件的配置
	Casbin struct {
		Disable             bool      // 是否禁用访问控制
		SkippedPathPrefixes []string  // 不需要访问控制的路径前缀列表
		LoadThread          int    `default:"2"`    // 加载策略的线程数
		AutoLoadInterval    int    `default:"3"`    // 自动重新加载策略的间隔（秒）
		ModelFile           string `default:"rbac_model.conf"`    // RBAC 模型配置文件路径
		GenPolicyFile       string `default:"gen_rbac_policy.csv"` // 生成的 RBAC 策略文件路径
	}
	// Static 定义了静态文件服务的配置
	Static struct {
		Dir string // 静态文件目录（从命令行参数获取）
	}
}
