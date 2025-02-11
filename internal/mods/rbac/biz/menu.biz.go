package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"
	"github.com/LyricTian/gin-admin/v10/pkg/encoding/yaml"
	"github.com/LyricTian/gin-admin/v10/pkg/errors"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"go.uber.org/zap"
)

// Menu 结构体定义了RBAC菜单管理所需的依赖
type Menu struct {
	Cache           cachex.Cacher     // 缓存接口
	Trans           *util.Trans       // 事务管理器
	MenuDAL         *dal.Menu         // 菜单数据访问层
	MenuResourceDAL *dal.MenuResource // 菜单资源数据访问层
	RoleMenuDAL     *dal.RoleMenu     // 角色菜单数据访问层
}

// InitFromFile 从配置文件初始化菜单数据
// 支持从JSON或YAML文件中读取菜单配置
func (a *Menu) InitFromFile(ctx context.Context, menuFile string) error {
	// 读取配置文件内容
	f, err := os.ReadFile(menuFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logging.Context(ctx).Warn("菜单数据文件未找到，跳过从文件初始化菜单数据", zap.String("file", menuFile))
			return nil
		}
		return err
	}

	var menus schema.Menus
	// 根据文件扩展名选择解析方式
	if ext := filepath.Ext(menuFile); ext == ".json" {
		if err := json.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "解析JSON文件 '%s' 失败", menuFile)
		}
	} else if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(f, &menus); err != nil {
			return errors.Wrapf(err, "解析YAML文件 '%s' 失败", menuFile)
		}
	} else {
		return errors.Errorf("不支持的文件类型 '%s'", ext)
	}

	// 在事务中执行菜单创建
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		return a.createInBatchByParent(ctx, menus, nil)
	})
}

// createInBatchByParent 批量创建菜单，支持父子层级关系
// items: 要创建的菜单项列表
// parent: 父菜单项，如果为nil则表示顶级菜单
func (a *Menu) createInBatchByParent(ctx context.Context, items schema.Menus, parent *schema.Menu) error {
	total := len(items)

	for i, item := range items {
		var parentID string
		if parent != nil {
			parentID = parent.ID
		}

		var (
			menuItem *schema.Menu
			err      error
		)

		// 尝试通过ID、Code或Name查找已存在的菜单
		if item.ID != "" {
			menuItem, err = a.MenuDAL.Get(ctx, item.ID)
		} else if item.Code != "" {
			menuItem, err = a.MenuDAL.GetByCodeAndParentID(ctx, item.Code, parentID)
		} else if item.Name != "" {
			menuItem, err = a.MenuDAL.GetByNameAndParentID(ctx, item.Name, parentID)
		}

		if err != nil {
			return err
		}

		// 设置默认状态为启用
		if item.Status == "" {
			item.Status = schema.MenuStatusEnabled
		}

		// 如果菜单已存在，检查并更新变更的字段
		if menuItem != nil {
			changed := false
			if menuItem.Name != item.Name {
				menuItem.Name = item.Name
				changed = true
			}
			if menuItem.Description != item.Description {
				menuItem.Description = item.Description
				changed = true
			}
			if menuItem.Path != item.Path {
				menuItem.Path = item.Path
				changed = true
			}
			if menuItem.Type != item.Type {
				menuItem.Type = item.Type
				changed = true
			}
			if menuItem.Sequence != item.Sequence {
				menuItem.Sequence = item.Sequence
				changed = true
			}
			if menuItem.Status != item.Status {
				menuItem.Status = item.Status
				changed = true
			}
			if changed {
				menuItem.UpdatedAt = time.Now()
				if err := a.MenuDAL.Update(ctx, menuItem); err != nil {
					return err
				}
			}
		} else {
			// 创建新菜单
			if item.ID == "" {
				item.ID = util.NewXID()
			}
			if item.Sequence == 0 {
				item.Sequence = total - i
			}
			item.ParentID = parentID
			if parent != nil {
				item.ParentPath = parent.ParentPath + parentID + util.TreePathDelimiter
			}
			menuItem = item
			if err := a.MenuDAL.Create(ctx, item); err != nil {
				return err
			}
		}

		// 处理菜单关联的资源
		for _, res := range item.Resources {
			// 检查资源是否已存在
			if res.ID != "" {
				exists, err := a.MenuResourceDAL.Exists(ctx, res.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}

			if res.Path != "" {
				exists, err := a.MenuResourceDAL.ExistsMethodPathByMenuID(ctx, res.Method, res.Path, menuItem.ID)
				if err != nil {
					return err
				} else if exists {
					continue
				}
			}
			if res.ID == "" {
				res.ID = util.NewXID()
			}
			res.MenuID = menuItem.ID
			if err := a.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		// 递归处理子菜单
		if item.Children != nil {
			if err := a.createInBatchByParent(ctx, *item.Children, menuItem); err != nil {
				return err
			}
		}
	}
	return nil
}

// Query 查询菜单列表
// params: 查询参数
func (a *Menu) Query(ctx context.Context, params schema.MenuQueryParam) (*schema.MenuQueryResult, error) {
	// 禁用分页
	params.Pagination = false

	// 填充查询参数
	if err := a.fillQueryParam(ctx, &params); err != nil {
		return nil, err
	}

	// 执行查询
	result, err := a.MenuDAL.Query(ctx, params, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			OrderFields: schema.MenusOrderParams,
		},
	})
	if err != nil {
		return nil, err
	}

	// 如果按名称模糊查询或按代码路径查询，需要追加子菜单
	if params.LikeName != "" || params.CodePath != "" {
		result.Data, err = a.appendChildren(ctx, result.Data)
		if err != nil {
			return nil, err
		}
	}

	// 如果需要包含资源信息，查询每个菜单的资源
	if params.IncludeResources {
		for i, item := range result.Data {
			resResult, err := a.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
				MenuID: item.ID,
			})
			if err != nil {
				return nil, err
			}
			result.Data[i].Resources = resResult.Data
		}
	}

	// 将结果转换为树形结构
	result.Data = result.Data.ToTree()
	return result, nil
}

// fillQueryParam 填充查询参数
// 主要处理代码路径相关的查询参数
func (a *Menu) fillQueryParam(ctx context.Context, params *schema.MenuQueryParam) error {
	if params.CodePath != "" {
		var (
			codes    []string
			lastMenu schema.Menu
		)
		// 解析代码路径，按层级查找菜单
		for _, code := range strings.Split(params.CodePath, util.TreePathDelimiter) {
			if code == "" {
				continue
			}
			codes = append(codes, code)
			menu, err := a.MenuDAL.GetByCodeAndParentID(ctx, code, lastMenu.ParentID, schema.MenuQueryOptions{
				QueryOptions: util.QueryOptions{
					SelectFields: []string{"id", "parent_id", "parent_path"},
				},
			})
			if err != nil {
				return err
			} else if menu == nil {
				return errors.NotFound("", "未找到代码为 '%s' 的菜单", strings.Join(codes, util.TreePathDelimiter))
			}
			lastMenu = *menu
		}
		params.ParentPathPrefix = lastMenu.ParentPath + lastMenu.ID + util.TreePathDelimiter
	}
	return nil
}

// appendChildren 追加子菜单数据
// 当查询结果需要包含子菜单时使用
func (a *Menu) appendChildren(ctx context.Context, data schema.Menus) (schema.Menus, error) {
	if len(data) == 0 {
		return data, nil
	}

	// 检查ID是否已存在于数据中
	existsInData := func(id string) bool {
		for _, item := range data {
			if item.ID == id {
				return true
			}
		}
		return false
	}

	// 查询并追加每个菜单的子菜单
	for _, item := range data {
		childResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
			ParentPathPrefix: item.ParentPath + item.ID + util.TreePathDelimiter,
		})
		if err != nil {
			return nil, err
		}
		for _, child := range childResult.Data {
			if existsInData(child.ID) {
				continue
			}
			data = append(data, child)
		}
	}

	// 查询并追加父菜单
	if parentIDs := data.SplitParentIDs(); len(parentIDs) > 0 {
		parentResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
			InIDs: parentIDs,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range parentResult.Data {
			if existsInData(p.ID) {
				continue
			}
			data = append(data, p)
		}
	}
	sort.Sort(data)

	return data, nil
}

// Get 获取指定ID的菜单详情
func (a *Menu) Get(ctx context.Context, id string) (*schema.Menu, error) {
	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return nil, err
	} else if menu == nil {
		return nil, errors.NotFound("", "菜单不存在")
	}

	// 查询菜单关联的资源
	menuResResult, err := a.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
		MenuID: menu.ID,
	})
	if err != nil {
		return nil, err
	}
	menu.Resources = menuResResult.Data

	return menu, nil
}

// Create 创建新菜单
func (a *Menu) Create(ctx context.Context, formItem *schema.MenuForm) (*schema.Menu, error) {
	// 检查是否允许操作菜单
	if config.C.General.DenyOperateMenu {
		return nil, errors.BadRequest("", "不允许创建菜单")
	}

	menu := &schema.Menu{
		ID:        util.NewXID(),
		CreatedAt: time.Now(),
	}

	// 处理父菜单关系
	if parentID := formItem.ParentID; parentID != "" {
		parent, err := a.MenuDAL.Get(ctx, parentID)
		if err != nil {
			return nil, err
		} else if parent == nil {
			return nil, errors.NotFound("", "父菜单不存在")
		}
		menu.ParentPath = parent.ParentPath + parent.ID + util.TreePathDelimiter
	}

	// 检查同级菜单下是否存在相同代码
	if exists, err := a.MenuDAL.ExistsCodeByParentID(ctx, formItem.Code, formItem.ParentID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.BadRequest("", "同级菜单下已存在相同代码")
	}

	// 填充表单数据到菜单对象
	if err := formItem.FillTo(menu); err != nil {
		return nil, err
	}

	// 在事务中创建菜单及其资源
	err := a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.MenuDAL.Create(ctx, menu); err != nil {
			return err
		}

		// 创建菜单关联的资源
		for _, res := range formItem.Resources {
			res.ID = util.NewXID()
			res.MenuID = menu.ID
			res.CreatedAt = time.Now()
			if err := a.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return menu, nil
}

// Update 更新指定菜单
func (a *Menu) Update(ctx context.Context, id string, formItem *schema.MenuForm) error {
	// 检查是否允许操作菜单
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "不允许更新菜单")
	}

	// 获取现有菜单
	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "菜单不存在")
	}

	// 保存原始的父路径和状态，用于后续处理
	oldParentPath := menu.ParentPath
	oldStatus := menu.Status
	var childData schema.Menus

	// 处理父菜单变更
	if menu.ParentID != formItem.ParentID {
		if parentID := formItem.ParentID; parentID != "" {
			parent, err := a.MenuDAL.Get(ctx, parentID)
			if err != nil {
				return err
			} else if parent == nil {
				return errors.NotFound("", "父菜单不存在")
			}
			menu.ParentPath = parent.ParentPath + parent.ID + util.TreePathDelimiter
		} else {
			menu.ParentPath = ""
		}

		// 查询所有子菜单，用于更新它们的父路径
		childResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
			ParentPathPrefix: oldParentPath + menu.ID + util.TreePathDelimiter,
		}, schema.MenuQueryOptions{
			QueryOptions: util.QueryOptions{
				SelectFields: []string{"id", "parent_path"},
			},
		})
		if err != nil {
			return err
		}
		childData = childResult.Data
	}

	// 检查菜单代码是否重复
	if menu.Code != formItem.Code {
		if exists, err := a.MenuDAL.ExistsCodeByParentID(ctx, formItem.Code, formItem.ParentID); err != nil {
			return err
		} else if exists {
			return errors.BadRequest("", "同级菜单下已存在相同代码")
		}
	}

	// 填充表单数据到菜单对象
	if err := formItem.FillTo(menu); err != nil {
		return err
	}

	// 在事务中执行更新操作
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		// 如果状态发生变化，更新所有子菜单的状态
		if oldStatus != formItem.Status {
			oldPath := oldParentPath + menu.ID + util.TreePathDelimiter
			if err := a.MenuDAL.UpdateStatusByParentPath(ctx, oldPath, formItem.Status); err != nil {
				return err
			}
		}

		// 更新子菜单的父路径
		for _, child := range childData {
			oldPath := oldParentPath + menu.ID + util.TreePathDelimiter
			newPath := menu.ParentPath + menu.ID + util.TreePathDelimiter
			err := a.MenuDAL.UpdateParentPath(ctx, child.ID, strings.Replace(child.ParentPath, oldPath, newPath, 1))
			if err != nil {
				return err
			}
		}

		// 更新菜单信息
		if err := a.MenuDAL.Update(ctx, menu); err != nil {
			return err
		}

		// 重新创建菜单资源
		if err := a.MenuResourceDAL.DeleteByMenuID(ctx, id); err != nil {
			return err
		}
		for _, res := range formItem.Resources {
			if res.ID == "" {
				res.ID = util.NewXID()
			}
			res.MenuID = id
			if res.CreatedAt.IsZero() {
				res.CreatedAt = time.Now()
			}
			res.UpdatedAt = time.Now()
			if err := a.MenuResourceDAL.Create(ctx, res); err != nil {
				return err
			}
		}

		return a.syncToCasbin(ctx)
	})
}

// Delete 删除指定菜单及其子菜单
func (a *Menu) Delete(ctx context.Context, id string) error {
	// 检查是否允许操作菜单
	if config.C.General.DenyOperateMenu {
		return errors.BadRequest("", "不允许删除菜单")
	}

	// 获取要删除的菜单
	menu, err := a.MenuDAL.Get(ctx, id)
	if err != nil {
		return err
	} else if menu == nil {
		return errors.NotFound("", "菜单不存在")
	}

	// 查询所有子菜单
	childResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
		ParentPathPrefix: menu.ParentPath + menu.ID + util.TreePathDelimiter,
	}, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id"},
		},
	})
	if err != nil {
		return err
	}

	// 在事务中执行删除操作
	return a.Trans.Exec(ctx, func(ctx context.Context) error {
		if err := a.delete(ctx, id); err != nil {
			return err
		}

		// 删除所有子菜单
		for _, child := range childResult.Data {
			if err := a.delete(ctx, child.ID); err != nil {
				return err
			}
		}

		return a.syncToCasbin(ctx)
	})
}

// delete 删除单个菜单及其关联数据
func (a *Menu) delete(ctx context.Context, id string) error {
	if err := a.MenuDAL.Delete(ctx, id); err != nil {
		return err
	}
	if err := a.MenuResourceDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	if err := a.RoleMenuDAL.DeleteByMenuID(ctx, id); err != nil {
		return err
	}
	return nil
}

// syncToCasbin 同步菜单变更到Casbin
// 通过缓存通知其他服务进行同步
func (a *Menu) syncToCasbin(ctx context.Context) error {
	return a.Cache.Set(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin, fmt.Sprintf("%d", time.Now().Unix()))
}
