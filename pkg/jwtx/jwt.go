// Package jwtx 提供了 JWT (JSON Web Token) 认证的实现
package jwtx

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

// Auther 定义了 JWT 认证的接口
type Auther interface {
	// GenerateToken 使用提供的主题（通常是用户标识）生成 JWT token
	GenerateToken(ctx context.Context, subject string) (TokenInfo, error)
	// DestroyToken 使指定的 token 失效
	DestroyToken(ctx context.Context, accessToken string) error
	// ParseSubject 从访问令牌中解析出主题（用户标识）
	ParseSubject(ctx context.Context, accessToken string) (string, error)
	// Release 释放 JWTAuth 实例持有的资源
	Release(ctx context.Context) error
}

// 默认的签名密钥
const defaultKey = "CG24SDVP8OHPK395GB5G"

// ErrInvalidToken 表示无效的 token 错误
var ErrInvalidToken = errors.New("Invalid token")

// options 存储 JWT 配置选项
type options struct {
	signingMethod jwt.SigningMethod                       // 签名方法
	signingKey    []byte                                  // 当前签名密钥
	signingKey2   []byte                                  // 旧的签名密钥（用于密钥轮换）
	keyFuncs      []func(*jwt.Token) (interface{}, error) // 用于验证 token 的密钥函数列表
	expired       int                                     // token 过期时间（秒）
	tokenType     string                                  // token 类型（如 "Bearer"）
}

// Option 定义了配置函数类型
type Option func(*options)

// SetSigningMethod 设置 JWT 的签名方法
func SetSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

// SetSigningKey 设置签名密钥，支持密钥轮换
func SetSigningKey(key, oldKey string) Option {
	return func(o *options) {
		o.signingKey = []byte(key)
		if oldKey != "" && key != oldKey {
			o.signingKey2 = []byte(oldKey)
		}
	}
}

// SetExpired 设置 token 的过期时间（秒）
func SetExpired(expired int) Option {
	return func(o *options) {
		o.expired = expired
	}
}

// New 创建一个新的 JWT 认证器
func New(store Storer, opts ...Option) Auther {
	// 设置默认选项
	o := options{
		tokenType:     "Bearer",
		expired:       7200, // 默认2小时过期
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    []byte(defaultKey),
	}

	// 应用自定义选项
	for _, opt := range opts {
		opt(&o)
	}

	// 添加主密钥的验证函数
	o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return o.signingKey, nil
	})

	// 如果存在旧密钥，添加旧密钥的验证函数
	if o.signingKey2 != nil {
		o.keyFuncs = append(o.keyFuncs, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return o.signingKey2, nil
		})
	}

	return &JWTAuth{
		opts:  &o,
		store: store,
	}
}

// JWTAuth 实现了 Auther 接口
type JWTAuth struct {
	opts  *options // JWT 配置选项
	store Storer   // token 存储器
}

// GenerateToken 生成新的 JWT token
func (a *JWTAuth) GenerateToken(ctx context.Context, subject string) (TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.opts.expired) * time.Second).Unix()

	// 创建 JWT 声明
	token := jwt.NewWithClaims(a.opts.signingMethod, &jwt.StandardClaims{
		IssuedAt:  now.Unix(), // 签发时间
		ExpiresAt: expiresAt,  // 过期时间
		NotBefore: now.Unix(), // 生效时间
		Subject:   subject,    // 主题（通常是用户ID）
	})

	// 签名并获取 token 字符串
	tokenStr, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		return nil, err
	}

	// 返回 token 信息
	tokenInfo := &tokenInfo{
		ExpiresAt:   expiresAt,
		TokenType:   a.opts.tokenType,
		AccessToken: tokenStr,
	}
	return tokenInfo, nil
}

// parseToken 解析 token 字符串
func (a *JWTAuth) parseToken(tokenStr string) (*jwt.StandardClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	// 尝试使用所有可用的密钥进行解析
	for _, keyFunc := range a.opts.keyFuncs {
		token, err = jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, keyFunc)
		if err != nil || token == nil || !token.Valid {
			continue
		}
		break
	}

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return token.Claims.(*jwt.StandardClaims), nil
}

// callStore 调用存储器的辅助函数
func (a *JWTAuth) callStore(fn func(Storer) error) error {
	if store := a.store; store != nil {
		return fn(store)
	}
	return nil
}

// DestroyToken 使 token 失效
func (a *JWTAuth) DestroyToken(ctx context.Context, tokenStr string) error {
	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return err
	}

	return a.callStore(func(store Storer) error {
		expired := time.Until(time.Unix(claims.ExpiresAt, 0))
		return store.Set(ctx, tokenStr, expired)
	})
}

// ParseSubject 从 token 中解析主题（用户标识）
func (a *JWTAuth) ParseSubject(ctx context.Context, tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", ErrInvalidToken
	}

	claims, err := a.parseToken(tokenStr)
	if err != nil {
		return "", err
	}

	// 检查 token 是否已被注销
	err = a.callStore(func(store Storer) error {
		if exists, err := store.Check(ctx, tokenStr); err != nil {
			return err
		} else if exists {
			return ErrInvalidToken
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}

// Release 释放资源
func (a *JWTAuth) Release(ctx context.Context) error {
	return a.callStore(func(store Storer) error {
		return store.Close(ctx)
	})
}
