// Package schema 定义了与登录相关的数据结构
package schema

import "strings"

// Captcha 验证码结构体
// 用于前后端交互时传递验证码信息
type Captcha struct {
	CaptchaID string `json:"captcha_id"` // 验证码ID，用于标识特定的验证码实例
}

// LoginForm 登录表单结构体
// 包含用户登录时需要提供的所有信息
type LoginForm struct {
	Username    string `json:"username" binding:"required"`     // 用户名，必填字段
	Password    string `json:"password" binding:"required"`     // 密码，必填字段（使用md5加密）
	CaptchaID   string `json:"captcha_id" binding:"required"`   // 验证码ID，必填字段，与前面的Captcha结构体对应
	CaptchaCode string `json:"captcha_code" binding:"required"` // 用户输入的验证码内容，必填字段
}

// Trim 方法用于处理登录表单中的字符串
// 去除用户名和验证码中的首尾空格，确保数据的清洁性
func (a *LoginForm) Trim() *LoginForm {
	a.Username = strings.TrimSpace(a.Username)
	a.CaptchaCode = strings.TrimSpace(a.CaptchaCode)
	return a
}

// UpdateLoginPassword 修改密码结构体
// 用于用户修改密码时的数据传输
type UpdateLoginPassword struct {
	OldPassword string `json:"old_password" binding:"required"` // 旧密码，必填字段（md5加密）
	NewPassword string `json:"new_password" binding:"required"` // 新密码，必填字段（md5加密）
}

// LoginToken 登录令牌结构体
// 用于存储用户登录成功后的认证信息
type LoginToken struct {
	AccessToken string `json:"access_token"` // 访问令牌，使用JWT（JSON Web Token）格式
	TokenType   string `json:"token_type"`   // 令牌类型，用于指定认证头的格式（例如：Authorization: Bearer <token>）
	ExpiresAt   int64  `json:"expires_at"`   // 令牌过期时间，使用Unix时间戳（秒）
}

// UpdateCurrentUser 更新当前用户信息的结构体
// 用于用户修改个人信息时的数据传输
type UpdateCurrentUser struct {
	Name   string `json:"name" binding:"required,max=64"` // 用户名称，必填，最大长度64个字符
	Phone  string `json:"phone" binding:"max=32"`         // 电话号码，可选，最大长度32个字符
	Email  string `json:"email" binding:"max=128"`        // 电子邮箱，可选，最大长度128个字符
	Remark string `json:"remark" binding:"max=1024"`      // 备注信息，可选，最大长度1024个字符
}
