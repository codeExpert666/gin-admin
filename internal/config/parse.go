// package config 提供应用程序配置的加载和解析功能
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LyricTian/gin-admin/v10/pkg/encoding/json"  // 用于解析 JSON 格式的配置文件
	"github.com/LyricTian/gin-admin/v10/pkg/encoding/toml"  // 用于解析 TOML 格式的配置文件
	"github.com/LyricTian/gin-admin/v10/pkg/errors"        // 错误处理包
	"github.com/creasty/defaults"                          // 用于设置结构体默认值
)

var (
	once sync.Once     // 确保配置只被加载一次的同步原语
	C    = new(Config) // 全局配置实例
)

// MustLoad 强制加载配置文件，如果加载失败则会panic
// dir: 配置文件所在的目录
// names: 配置文件名列表
func MustLoad(dir string, names ...string) {
	once.Do(func() { // 使用 sync.Once 确保配置只被加载一次
		if err := Load(dir, names...); err != nil {
			panic(err)
		}
	})
}

// Load 从指定目录加载配置文件并解析到结构体中
// dir: 配置文件所在的目录
// names: 配置文件名列表
// 支持的配置文件格式：JSON、TOML
func Load(dir string, names ...string) error {
	// 首先设置配置结构体的默认值
	if err := defaults.Set(C); err != nil {
		return err
	}

	// 定义支持的配置文件扩展名
	supportExts := []string{".json", ".toml"}

	// parseFile 是一个内部函数，用于解析单个配置文件
	parseFile := func(name string) error {
		// 获取文件扩展名
		ext := filepath.Ext(name)
		// 如果文件没有扩展名或扩展名不支持，则跳过
		if ext == "" || !strings.Contains(strings.Join(supportExts, ","), ext) {
			return nil
		}

		// 读取配置文件内容
		buf, err := os.ReadFile(name)
		if err != nil {
			return errors.Wrapf(err, "failed to read config file %s", name)
		}

		// 根据文件扩展名选择相应的解析器
		switch ext {
		case ".json":
			err = json.Unmarshal(buf, C)  // 解析 JSON 格式
		case ".toml":
			err = toml.Unmarshal(buf, C)  // 解析 TOML 格式
		}
		// 如果解析失败，包装错误信息并返回
		return errors.Wrapf(err, "failed to unmarshal config %s", name)
	}

	// 遍历所有配置文件名
	for _, name := range names {
		// 构建完整的文件路径
		fullname := filepath.Join(dir, name)
		// 获取文件信息
		info, err := os.Stat(fullname)
		if err != nil {
			return errors.Wrapf(err, "failed to get config file %s", name)
		}

		// 如果是目录，则递归遍历目录下的所有文件
		if info.IsDir() {
			err := filepath.WalkDir(fullname, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				} else if d.IsDir() { // 跳过子目录
					return nil
				}
				// 解析目录下的每个文件
				return parseFile(path)
			})
			if err != nil {
				return errors.Wrapf(err, "failed to walk config dir %s", name)
			}
			continue
		}
		// 如果是文件，直接解析
		if err := parseFile(fullname); err != nil {
			return err
		}
	}

	return nil
}
