// Package gormx 提供了 GORM 数据库的配置和初始化功能
package gormx

import (
	// 导入所需的标准库和第三方包
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdmysql "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

// ResolverConfig 定义数据库读写分离的配置结构
// 读写分离的主要目的是将数据库读写操作分散到不同的数据库实例上，以提高性能：
// - 主库负责写操作（INSERT、UPDATE、DELETE）
// - 从库负责读操作（SELECT）
// - 通过数据库主从复制保持数据同步
type ResolverConfig struct {
	DBType   string   // 数据库类型：支持 mysql/postgres/sqlite3
	Sources  []string // 写库（主库）的连接地址列表
	Replicas []string // 读库（从库）的连接地址列表
	Tables   []string // 需要进行读写分离的数据库表名列表
}

// Config 定义 GORM 数据库的主要配置结构
type Config struct {
	Debug        bool             // 是否开启调试模式，开启后会打印详细的 SQL 日志
	PrepareStmt  bool             // 是否启用 prepared statement 缓存，可提高性能
	DBType       string           // 数据库类型：支持 mysql/postgres/sqlite3
	DSN          string           // Database Source Name，数据库连接字符串
	MaxLifetime  int              // 连接的最大生命周期（秒）
	MaxIdleTime  int              // 空闲连接的最大生命周期（秒）
	MaxOpenConns int              // 数据库连接池最大连接数
	MaxIdleConns int              // 数据库连接池最大空闲连接数
	TablePrefix  string           // 数据库表名前缀
	Resolver     []ResolverConfig // 读写分离配置列表
}

// New 创建并初始化一个新的 GORM 数据库实例
func New(cfg Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 根据配置的数据库类型选择对应的数据库驱动
	switch strings.ToLower(cfg.DBType) {
	case "mysql":
		// MySQL 数据库：如果数据库不存在则自动创建
		if err := createDatabaseWithMySQL(cfg.DSN); err != nil {
			return nil, err
		}
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite3":
		// SQLite 数据库：自动创建数据库文件的目录
		_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	// 配置 GORM 的基本参数
	ormCfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix, // 设置表名前缀
			SingularTable: true,            // 使用单数表名
		},
		Logger:      logger.Discard,  // 默认不打印日志
		PrepareStmt: cfg.PrepareStmt, // 是否启用 prepared statement
	}

	// 如果开启调试模式，则使用默认日志记录器
	if cfg.Debug {
		ormCfg.Logger = logger.Default
	}

	// 创建 GORM 数据库实例
	db, err := gorm.Open(dialector, ormCfg)
	if err != nil {
		return nil, err
	}

	// 配置读写分离
	if len(cfg.Resolver) > 0 {
		resolver := &dbresolver.DBResolver{}
		for _, r := range cfg.Resolver {
			resolverCfg := dbresolver.Config{}
			var open func(dsn string) gorm.Dialector
			dbType := strings.ToLower(r.DBType)

			// 选择对应的数据库驱动
			switch dbType {
			case "mysql":
				open = mysql.Open
			case "postgres":
				open = postgres.Open
			case "sqlite3":
				open = sqlite.Open
			default:
				continue
			}

			// 配置从库连接
			for _, replica := range r.Replicas {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
				}
				resolverCfg.Replicas = append(resolverCfg.Replicas, open(replica))
			}

			// 配置主库连接
			for _, source := range r.Sources {
				if dbType == "sqlite3" {
					_ = os.MkdirAll(filepath.Dir(cfg.DSN), os.ModePerm)
				}
				resolverCfg.Sources = append(resolverCfg.Sources, open(source))
			}

			// 注册读写分离配置
			tables := stringSliceToInterfaceSlice(r.Tables)
			resolver.Register(resolverCfg, tables...)
			zap.L().Info(fmt.Sprintf("Use resolver, #tables: %v, #replicas: %v, #sources: %v \n",
				tables, r.Replicas, r.Sources))
		}

		// 设置连接池参数
		resolver.SetMaxIdleConns(cfg.MaxIdleConns).
			SetMaxOpenConns(cfg.MaxOpenConns).
			SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second).
			SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)
		if err := db.Use(resolver); err != nil {
			return nil, err
		}
	}

	// 如果开启调试模式，启用 GORM 的调试模式
	if cfg.Debug {
		db = db.Debug()
	}

	// 获取底层的 *sql.DB 对象并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)                                // 设置空闲连接池的最大连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)                                // 设置数据库连接池的最大连接数
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second) // 设置连接的最大生命周期
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second) // 设置空闲连接的最大生命周期

	return db, nil
}

// stringSliceToInterfaceSlice 辅助函数：将字符串切片转换为接口切片
func stringSliceToInterfaceSlice(s []string) []interface{} {
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}

// createDatabaseWithMySQL 辅助函数：如果 MySQL 数据库不存在则自动创建
func createDatabaseWithMySQL(dsn string) error {
	// 解析 DSN（数据库连接字符串）
	cfg, err := sdmysql.ParseDSN(dsn)
	if err != nil {
		return err
	}

	// 创建一个到 MySQL 服务器的连接（不指定数据库）
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr))
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建数据库（如果不存在），并设置默认字符集为 utf8mb4
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`;", cfg.DBName)
	_, err = db.Exec(query)
	return err
}
