// Package cachex 实现了基于 Badger 的缓存系统
package cachex

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/dgraph-io/badger/v3"
)

// BadgerConfig 定义 Badger 数据库的配置结构
type BadgerConfig struct {
	Path string // 数据库存储路径
}

// NewBadgerCache 创建一个新的基于 Badger 的缓存实例
// cfg: Badger 配置
// opts: 可选的配置选项
func NewBadgerCache(cfg BadgerConfig, opts ...Option) Cacher {
	// 设置默认选项
	defaultOpts := &options{
		Delimiter: defaultDelimiter,
	}

	// 应用自定义选项
	for _, o := range opts {
		o(defaultOpts)
	}

	// 初始化 Badger 数据库配置
	badgerOpts := badger.DefaultOptions(cfg.Path)
	badgerOpts = badgerOpts.WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(badgerOpts)
	if err != nil {
		panic(err)
	}

	return &badgerCache{
		opts: defaultOpts,
		db:   db,
	}
}

// badgerCache 实现了 Cacher 接口的具体缓存结构
type badgerCache struct {
	opts *options   // 缓存配置选项
	db   *badger.DB // Badger 数据库实例
}

// getKey 生成缓存键，将命名空间和键名组合
func (a *badgerCache) getKey(ns, key string) string {
	return fmt.Sprintf("%s%s%s", ns, a.opts.Delimiter, key)
}

// strToBytes 使用 unsafe 包高效地将字符串转换为字节切片
// 注意：这是一个性能优化，但需要谨慎使用
func (a *badgerCache) strToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// bytesToStr 使用 unsafe 包高效地将字节切片转换为字符串
func (a *badgerCache) bytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Set 设置缓存键值对
// ns: 命名空间
// key: 键名
// value: 值
// expiration: 可选的过期时间
func (a *badgerCache) Set(ctx context.Context, ns, key, value string, expiration ...time.Duration) error {
	return a.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(a.strToBytes(a.getKey(ns, key)), a.strToBytes(value))
		if len(expiration) > 0 {
			entry = entry.WithTTL(expiration[0])
		}
		return txn.SetEntry(entry)
	})
}

// Get 获取缓存值
// 返回值: 缓存的值，是否存在，错误信息
func (a *badgerCache) Get(ctx context.Context, ns, key string) (string, bool, error) {
	value := ""
	ok := false
	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(a.strToBytes(a.getKey(ns, key)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		ok = true
		val, err := item.ValueCopy(nil)
		value = a.bytesToStr(val)
		return err
	})
	if err != nil {
		return "", false, err
	}
	return value, ok, nil
}

// Exists 检查键是否存在
func (a *badgerCache) Exists(ctx context.Context, ns, key string) (bool, error) {
	exists := false
	err := a.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(a.strToBytes(a.getKey(ns, key)))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		exists = true
		return nil
	})
	return exists, err
}

// Delete 删除缓存键
func (a *badgerCache) Delete(ctx context.Context, ns, key string) error {
	b, err := a.Exists(ctx, ns, key)
	if err != nil {
		return err
	} else if !b {
		return nil
	}

	return a.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(a.strToBytes(a.getKey(ns, key)))
	})
}

// GetAndDelete 获取缓存值并删除
// 原子操作：获取值后立即删除
func (a *badgerCache) GetAndDelete(ctx context.Context, ns, key string) (string, bool, error) {
	value, ok, err := a.Get(ctx, ns, key)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}

	err = a.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(a.strToBytes(a.getKey(ns, key)))
	})
	if err != nil {
		return "", false, err
	}

	return value, true, nil
}

// Iterator 遍历指定命名空间下的所有键值对
// fn: 回调函数，返回 false 时停止遍历
func (a *badgerCache) Iterator(ctx context.Context, ns string, fn func(ctx context.Context, key, value string) bool) error {
	return a.db.View(func(txn *badger.Txn) error {
		iterOpts := badger.DefaultIteratorOptions
		iterOpts.Prefix = a.strToBytes(a.getKey(ns, ""))
		it := txn.NewIterator(iterOpts)
		defer it.Close()

		it.Rewind()
		for it.Valid() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			key := a.bytesToStr(item.Key())
			if !fn(ctx, strings.TrimPrefix(key, a.getKey(ns, "")), a.bytesToStr(val)) {
				break
			}
			it.Next()
		}
		return nil
	})
}

// Close 关闭缓存，清理资源
func (a *badgerCache) Close(ctx context.Context) error {
	return a.db.Close()
}
