package config

// 缓存命名空间常量定义
const (
	// CacheNSForUser 用户相关的缓存命名空间
	// 用于存储用户信息、权限等数据
	CacheNSForUser = "user"

	// CacheNSForRole 角色相关的缓存命名空间
	// 用于存储角色信息、权限等数据
	CacheNSForRole = "role"
)

// Casbin 相关的缓存键定义
const (
	// CacheKeyForSyncToCasbin Casbin 同步标记的缓存键
	// 当角色权限发生变化时，通过此键通知系统同步 Casbin 权限规则
	CacheKeyForSyncToCasbin = "sync:casbin"
)

// 系统错误码常量定义
const (
	// ErrInvalidTokenID 无效的令牌错误
	// 当用户提供的访问令牌无效或已过期时返回此错误
	ErrInvalidTokenID = "com.invalid.token"

	// ErrInvalidCaptchaID 无效的验证码错误
	// 当用户提供的验证码不正确或已过期时返回此错误
	ErrInvalidCaptchaID = "com.invalid.captcha"

	// ErrInvalidUsernameOrPassword 用户名或密码错误
	// 当用户登录时提供的用户名或密码不正确时返回此错误
	ErrInvalidUsernameOrPassword = "com.invalid.username-or-password"
)
