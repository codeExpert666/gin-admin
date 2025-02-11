// Package rbac 实现基于角色的访问控制
package rbac

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LyricTian/gin-admin/v10/internal/config"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/dal"
	"github.com/LyricTian/gin-admin/v10/internal/mods/rbac/schema"
	"github.com/LyricTian/gin-admin/v10/pkg/cachex"
	"github.com/LyricTian/gin-admin/v10/pkg/logging"
	"github.com/LyricTian/gin-admin/v10/pkg/util"
	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
)

// Casbinx 结构体用于加载和管理 RBAC 权限
type Casbinx struct {
	enforcer        *atomic.Value     `wire:"-"` // 原子值存储 casbin enforcer，确保线程安全
	ticker          *time.Ticker      `wire:"-"` // 用于定时自动加载策略的计时器
	Cache           cachex.Cacher     // 缓存接口
	MenuDAL         *dal.Menu         // 菜单数据访问层
	MenuResourceDAL *dal.MenuResource // 菜单资源数据访问层
	RoleDAL         *dal.Role         // 角色数据访问层
}

// GetEnforcer 获取当前的 casbin enforcer 实例
func (a *Casbinx) GetEnforcer() *casbin.Enforcer {
	if v := a.enforcer.Load(); v != nil {
		return v.(*casbin.Enforcer)
	}
	return nil
}

// policyQueueItem 定义策略队列项，用于并发处理策略加载
type policyQueueItem struct {
	RoleID    string               // 角色ID
	Resources schema.MenuResources // 角色对应的资源列表
}

// Load 初始化并加载 RBAC 权限
func (a *Casbinx) Load(ctx context.Context) error {
	// 如果在配置中禁用了 Casbin，直接返回
	if config.C.Middleware.Casbin.Disable {
		return nil
	}

	// 初始化 enforcer 原子值
	a.enforcer = new(atomic.Value)
	if err := a.load(ctx); err != nil {
		return err
	}

	// 启动自动加载策略的后台任务
	go a.autoLoad(ctx)
	return nil
}

// load 执行实际的策略加载工作
func (a *Casbinx) load(ctx context.Context) error {
	start := time.Now()
	// 查询所有启用状态的角色
	roleResult, err := a.RoleDAL.Query(ctx, schema.RoleQueryParam{
		Status: schema.RoleStatusEnabled,
	}, schema.RoleQueryOptions{
		QueryOptions: util.QueryOptions{SelectFields: []string{"id"}},
	})
	if err != nil {
		return err
	} else if len(roleResult.Data) == 0 {
		return nil
	}

	var resCount int32                                         // 资源计数器
	queue := make(chan *policyQueueItem, len(roleResult.Data)) // 创建策略处理队列
	threadNum := config.C.Middleware.Casbin.LoadThread         // 获取配置的线程数
	lock := new(sync.Mutex)                                    // 创建互斥锁
	buf := new(bytes.Buffer)                                   // 创建缓冲区存储策略

	// 创建工作组处理策略加载
	wg := new(sync.WaitGroup)
	wg.Add(threadNum)
	// 启动多个 goroutine 并发处理策略
	for i := 0; i < threadNum; i++ {
		go func() {
			defer wg.Done()
			ibuf := new(bytes.Buffer)
			// 从队列中获取策略项并处理
			for item := range queue {
				for _, res := range item.Resources {
					_, _ = ibuf.WriteString(fmt.Sprintf("p, %s, %s, %s \n", item.RoleID, res.Path, res.Method))
				}
			}
			// 将处理结果写入主缓冲区
			lock.Lock()
			_, _ = buf.Write(ibuf.Bytes())
			lock.Unlock()
		}()
	}

	// 遍历角色，查询其资源并加入处理队列
	for _, item := range roleResult.Data {
		resources, err := a.queryRoleResources(ctx, item.ID)
		if err != nil {
			logging.Context(ctx).Error("Failed to query role resources", zap.Error(err))
			continue
		}
		atomic.AddInt32(&resCount, int32(len(resources)))
		queue <- &policyQueueItem{
			RoleID:    item.ID,
			Resources: resources,
		}
	}
	close(queue)
	wg.Wait()

	// 如果有策略数据，写入文件并创建新的 enforcer
	if buf.Len() > 0 {
		policyFile := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.GenPolicyFile)
		_ = os.Rename(policyFile, policyFile+".bak") // 备份原策略文件
		_ = os.MkdirAll(filepath.Dir(policyFile), 0755)
		if err := os.WriteFile(policyFile, buf.Bytes(), 0666); err != nil {
			logging.Context(ctx).Error("Failed to write policy file", zap.Error(err))
			return err
		}
		// 设置文件为只读
		_ = os.Chmod(policyFile, 0444)

		// 创建新的 enforcer
		modelFile := filepath.Join(config.C.General.WorkDir, config.C.Middleware.Casbin.ModelFile)
		e, err := casbin.NewEnforcer(modelFile, policyFile)
		if err != nil {
			logging.Context(ctx).Error("Failed to create casbin enforcer", zap.Error(err))
			return err
		}
		e.EnableLog(config.C.IsDebug())
		a.enforcer.Store(e)
	}

	// 记录加载统计信息
	logging.Context(ctx).Info("Casbin load policy",
		zap.Duration("cost", time.Since(start)),
		zap.Int("roles", len(roleResult.Data)),
		zap.Int32("resources", resCount),
		zap.Int("bytes", buf.Len()),
	)
	return nil
}

// queryRoleResources 查询指定角色ID的所有资源
func (a *Casbinx) queryRoleResources(ctx context.Context, roleID string) (schema.MenuResources, error) {
	// 查询角色关联的所有启用状态的菜单
	menuResult, err := a.MenuDAL.Query(ctx, schema.MenuQueryParam{
		RoleID: roleID,
		Status: schema.MenuStatusEnabled,
	}, schema.MenuQueryOptions{
		QueryOptions: util.QueryOptions{
			SelectFields: []string{"id", "parent_id", "parent_path"},
		},
	})
	if err != nil {
		return nil, err
	} else if len(menuResult.Data) == 0 {
		return nil, nil
	}

	// 收集所有相关的菜单ID（包括父菜单）
	menuIDs := make([]string, 0, len(menuResult.Data))
	menuIDMapper := make(map[string]struct{})
	for _, item := range menuResult.Data {
		if _, ok := menuIDMapper[item.ID]; ok {
			continue
		}
		menuIDs = append(menuIDs, item.ID)
		menuIDMapper[item.ID] = struct{}{}
		// 处理父菜单路径
		if pp := item.ParentPath; pp != "" {
			for _, pid := range strings.Split(pp, util.TreePathDelimiter) {
				if pid == "" {
					continue
				}
				if _, ok := menuIDMapper[pid]; ok {
					continue
				}
				menuIDs = append(menuIDs, pid)
				menuIDMapper[pid] = struct{}{}
			}
		}
	}

	// 查询这些菜单关联的资源
	menuResourceResult, err := a.MenuResourceDAL.Query(ctx, schema.MenuResourceQueryParam{
		MenuIDs: menuIDs,
	})
	if err != nil {
		return nil, err
	}

	return menuResourceResult.Data, nil
}

// autoLoad 自动定期加载策略的后台任务
func (a *Casbinx) autoLoad(ctx context.Context) {
	var lastUpdated int64 // 记录最后更新时间
	// 创建定时器，间隔时间从配置中获取
	a.ticker = time.NewTicker(time.Duration(config.C.Middleware.Casbin.AutoLoadInterval) * time.Second)
	for range a.ticker.C {
		// 从缓存中获取同步标记
		val, ok, err := a.Cache.Get(ctx, config.CacheNSForRole, config.CacheKeyForSyncToCasbin)
		if err != nil {
			logging.Context(ctx).Error("Failed to get cache", zap.Error(err), zap.String("key", config.CacheKeyForSyncToCasbin))
			continue
		} else if !ok {
			continue
		}

		// 解析更新时间
		updated, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logging.Context(ctx).Error("Failed to parse cache value", zap.Error(err), zap.String("val", val))
			continue
		}

		// 如果有更新，重新加载策略
		if lastUpdated < updated {
			if err := a.load(ctx); err != nil {
				logging.Context(ctx).Error("Failed to load casbin policy", zap.Error(err))
			} else {
				lastUpdated = updated
			}
		}
	}
}

// Release 释放资源
func (a *Casbinx) Release(ctx context.Context) error {
	if a.ticker != nil {
		a.ticker.Stop()
	}
	return nil
}
