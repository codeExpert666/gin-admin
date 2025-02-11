package biz

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/LyricTian/captcha"
	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/crypto/hash"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/jwtx"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Login 结构体用于处理RBAC（基于角色的访问控制）中的登录管理
// 包含了缓存、认证、用户数据访问等必要组件
type Login struct {
	Cache       cachex.Cacher // 缓存接口，用于存储用户会话信息
	Auth        jwtx.Auther   // JWT认证接口
	UserDAL     *dal.User     // 用户数据访问层
	UserRoleDAL *dal.UserRole // 用户角色数据访问层
	MenuDAL     *dal.Menu     // 菜单数据访问层
	UserBIZ     *User         // 用户业务逻辑层
}

// ParseUserID 从请求上下文中解析用户ID
// 主要用于验证用户身份和获取用户信息
func (a *Login) ParseUserID(c *gin.Context) (string, error) {
	rootID := config.C.General.Root.ID
	// 如果禁用了认证中间件，直接返回root用户ID
	if config.C.Middleware.Auth.Disable {
		return rootID, nil
	}

	invalidToken := errors.Unauthorized(config.ErrInvalidTokenID, "Invalid access token")
	// 从请求中获取token
	token := util.GetToken(c)
	if token == "" {
		return "", invalidToken
	}

	ctx := c.Request.Context()
	ctx = util.NewUserToken(ctx, token)

	// 解析token中的用户ID
	userID, err := a.Auth.ParseSubject(ctx, token)
	if err != nil {
		if err == jwtx.ErrInvalidToken {
			return "", invalidToken
		}
		return "", err
	} else if userID == rootID {
		// 如果是root用户，设置特殊标记
		c.Request = c.Request.WithContext(util.NewIsRootUser(ctx))
		return userID, nil
	}

	// 从缓存中获取用户信息
	userCacheVal, ok, err := a.Cache.Get(ctx, config.CacheNSForUser, userID)
	if err != nil {
		return "", err
	} else if ok {
		userCache := util.ParseUserCache(userCacheVal)
		c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
		return userID, nil
	}

	// 检查用户状态，如果未激活则强制登出
	user, err := a.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{SelectFields: []string{"status"}},
	})
	if err != nil {
		return "", err
	} else if user == nil || user.Status != schema.UserStatusActivated {
		return "", invalidToken
	}

	// 获取用户角色ID列表
	roleIDs, err := a.UserBIZ.GetRoleIDs(ctx, userID)
	if err != nil {
		return "", err
	}

	// 将用户信息存入缓存
	userCache := util.UserCache{
		RoleIDs: roleIDs,
	}
	err = a.Cache.Set(ctx, config.CacheNSForUser, userID, userCache.String())
	if err != nil {
		return "", err
	}

	c.Request = c.Request.WithContext(util.NewUserCache(ctx, userCache))
	return userID, nil
}

// GetCaptcha 生成新的验证码
// 返回验证码ID，验证码长度由配置决定
func (a *Login) GetCaptcha(ctx context.Context) (*schema.Captcha, error) {
	return &schema.Captcha{
		CaptchaID: captcha.NewLen(config.C.Util.Captcha.Length),
	}, nil
}

// ResponseCaptcha 响应验证码图片
// 生成验证码图片并设置相应的HTTP头
func (a *Login) ResponseCaptcha(ctx context.Context, w http.ResponseWriter, id string, reload bool) error {
	if reload && !captcha.Reload(id) {
		return errors.NotFound("", "Captcha id not found")
	}

	err := captcha.WriteImage(w, id, config.C.Util.Captcha.Width, config.C.Util.Captcha.Height)
	if err != nil {
		if err == captcha.ErrNotFound {
			return errors.NotFound("", "Captcha id not found")
		}
		return err
	}

	// 设置HTTP响应头，确保验证码图片不被缓存
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "image/png")
	return nil
}

// genUserToken 生成用户访问令牌
// 包含访问令牌、令牌类型和过期时间
func (a *Login) genUserToken(ctx context.Context, userID string) (*schema.LoginToken, error) {
	token, err := a.Auth.GenerateToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokenBuf, err := token.EncodeToJSON()
	if err != nil {
		return nil, err
	}
	logging.Context(ctx).Info("Generate user token", zap.Any("token", string(tokenBuf)))

	return &schema.LoginToken{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
		ExpiresAt:   token.GetExpiresAt(),
	}, nil
}

// Login 处理用户登录请求
// 验证验证码、用户名和密码，成功后生成访问令牌
func (a *Login) Login(ctx context.Context, formItem *schema.LoginForm) (*schema.LoginToken, error) {
	// 验证验证码
	if !captcha.VerifyString(formItem.CaptchaID, formItem.CaptchaCode) {
		return nil, errors.BadRequest(config.ErrInvalidCaptchaID, "Incorrect captcha")
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogin)

	// 处理root用户登录
	if formItem.Username == config.C.General.Root.Username {
		if formItem.Password != config.C.General.Root.Password {
			return nil, errors.BadRequest(config.ErrInvalidUsernameOrPassword, "Incorrect username or password")
		}

		userID := config.C.General.Root.ID
		ctx = logging.NewUserID(ctx, userID)
		logging.Context(ctx).Info("Login by root")
		return a.genUserToken(ctx, userID)
	}

	// 获取普通用户信息
	user, err := a.UserDAL.GetByUsername(ctx, formItem.Username, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id", "password", "status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.BadRequest(config.ErrInvalidUsernameOrPassword, "Incorrect username or password")
	} else if user.Status != schema.UserStatusActivated {
		return nil, errors.BadRequest("", "User status is not activated, please contact the administrator")
	}

	// 验证密码
	if err := hash.CompareHashAndPassword(user.Password, formItem.Password); err != nil {
		return nil, errors.BadRequest(config.ErrInvalidUsernameOrPassword, "Incorrect username or password")
	}

	userID := user.ID
	ctx = logging.NewUserID(ctx, userID)

	// 设置用户缓存和角色信息
	roleIDs, err := a.UserBIZ.GetRoleIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	userCache := util.UserCache{RoleIDs: roleIDs}
	err = a.Cache.Set(ctx, config.CacheNSForUser, userID, userCache.String(),
		time.Duration(config.C.Dictionary.UserCacheExp)*time.Hour)
	if err != nil {
		logging.Context(ctx).Error("Failed to set cache", zap.Error(err))
	}
	logging.Context(ctx).Info("Login success", zap.String("username", formItem.Username))

	// 生成访问令牌
	return a.genUserToken(ctx, userID)
}

// RefreshToken 刷新用户访问令牌
func (a *Login) RefreshToken(ctx context.Context) (*schema.LoginToken, error) {
	userID := util.FromUserID(ctx)

	// 检查用户状态
	user, err := a.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"status"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.BadRequest("", "Incorrect user")
	} else if user.Status != schema.UserStatusActivated {
		return nil, errors.BadRequest("", "User status is not activated, please contact the administrator")
	}

	return a.genUserToken(ctx, userID)
}

// Logout 处理用户登出请求
// 销毁令牌并清除用户缓存
func (a *Login) Logout(ctx context.Context) error {
	userToken := util.FromUserToken(ctx)
	if userToken == "" {
		return nil
	}

	ctx = logging.NewTag(ctx, logging.TagKeyLogout)
	if err := a.Auth.DestroyToken(ctx, userToken); err != nil {
		return err
	}

	userID := util.FromUserID(ctx)
	err := a.Cache.Delete(ctx, config.CacheNSForUser, userID)
	if err != nil {
		logging.Context(ctx).Error("Failed to delete user cache", zap.Error(err))
	}
	logging.Context(ctx).Info("Logout success")

	return nil
}

// GetUserInfo 获取用户信息
// 包括用户基本信息和角色信息
func (a *Login) GetUserInfo(ctx context.Context) (*schema.User, error) {
	if util.FromIsRootUser(ctx) {
		return &schema.User{
			ID:       config.C.General.Root.ID,
			Username: config.C.General.Root.Username,
			Name:     config.C.General.Root.Name,
			Status:   schema.UserStatusActivated,
		}, nil
	}

	userID := util.FromUserID(ctx)
	user, err := a.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			OmitFields: []string{"password"},
		},
	})
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, errors.NotFound("", "User not found")
	}

	// 获取用户角色信息
	userRoleResult, err := a.UserRoleDAL.Query(ctx, schema.UserRoleQueryParam{
		UserID: userID,
	}, schema.UserRoleQueryOptions{
		JoinRole: true,
	})
	if err != nil {
		return nil, err
	}
	user.Roles = userRoleResult.Data

	return user, nil
}

// UpdatePassword 修改用户登录密码
func (a *Login) UpdatePassword(ctx context.Context, updateItem *schema.UpdateLoginPassword) error {
	if util.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot change password")
	}

	userID := util.FromUserID(ctx)
	user, err := a.UserDAL.Get(ctx, userID, schema.UserQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"password"},
		},
	})
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// 验证旧密码
	if err := hash.CompareHashAndPassword(user.Password, updateItem.OldPassword); err != nil {
		return errors.BadRequest("", "Incorrect old password")
	}

	// 更新新密码
	newPassword, err := hash.GeneratePassword(updateItem.NewPassword)
	if err != nil {
		return err
	}
	return a.UserDAL.UpdatePasswordByID(ctx, userID, newPassword)
}

// QueryMenus 查询用户可访问的菜单
// 根据用户权限返回菜单树结构
func (a *Login) QueryMenus(ctx context.Context) (schema.Menus, error) {
	menuQueryParams := schema.MenuQueryParam{
		Status: schema.MenuStatusEnabled,
	}

	isRoot := util.FromIsRootUser(ctx)
	if !isRoot {
		menuQueryParams.UserID = util.FromUserID(ctx)
	}

	// 查询菜单数据
	menuResult, err := a.MenuDAL.Query(ctx, menuQueryParams, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	} else if isRoot {
		return menuResult.Data.ToTree(), nil
	}

	// 填充父级菜单
	if parentIDs := menuResult.Data.SplitParentIDs(); len(parentIDs) > 0 {
		var missMenusIDs []string
		menuIDMapper := menuResult.Data.ToMap()
		for _, parentID := range parentIDs {
			if _, ok := menuIDMapper[parentID]; !ok {
				missMenusIDs = append(missMenusIDs, parentID)
			}
		}
		if len(missMenusIDs) > 0 {
			parentResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
				InIDs: missMenusIDs,
			})
			if err != nil {
				return nil, err
			}
			menuResult.Data = append(menuResult.Data, parentResult.Data...)
			sort.Sort(menuResult.Data)
		}
	}

	return menuResult.Data.ToTree(), nil
}

// UpdateUser 更新当前用户信息
func (a *Login) UpdateUser(ctx context.Context, updateItem *schema.UpdateCurrentUser) error {
	if util.FromIsRootUser(ctx) {
		return errors.BadRequest("", "Root user cannot update")
	}

	userID := util.FromUserID(ctx)
	user, err := a.UserDAL.Get(ctx, userID)
	if err != nil {
		return err
	} else if user == nil {
		return errors.NotFound("", "User not found")
	}

	// 更新用户基本信息
	user.Name = updateItem.Name
	user.Phone = updateItem.Phone
	user.Email = updateItem.Email
	user.Remark = updateItem.Remark
	return a.UserDAL.Update(ctx, user, "name", "phone", "email", "remark")
}
