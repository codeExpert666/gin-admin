// Package util 提供了一系列实用工具函数
package util

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Trans 定义了数据库事务的结构体
type Trans struct {
	DB *gorm.DB // 保存gorm数据库连接实例
}

// TransFunc 定义了事务执行函数的类型
type TransFunc func(context.Context) error

// Exec 执行数据库事务
// 如果上下文中已经存在事务,则直接执行;否则创建新的事务
func (a *Trans) Exec(ctx context.Context, fn TransFunc) error {
	// 检查上下文中是否已存在事务
	if _, ok := FromTrans(ctx); ok {
		return fn(ctx)
	}

	// 创建新的事务并执行
	return a.DB.Transaction(func(db *gorm.DB) error {
		return fn(NewTrans(ctx, db))
	})
}

// GetDB 获取数据库连接实例
// 根据上下文返回合适的数据库连接,支持事务和行锁
func GetDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	db := defDB
	// 如果上下文中存在事务,使用事务的数据库连接
	if tdb, ok := FromTrans(ctx); ok {
		db = tdb
	}
	// 如果需要行锁,添加FOR UPDATE子句
	if FromRowLock(ctx) {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return db.WithContext(ctx)
}

// wrapQueryOptions 包装查询选项
// 处理字段选择、字段忽略和排序等查询选项
func wrapQueryOptions(db *gorm.DB, opts QueryOptions) *gorm.DB {
	// 设置要查询的字段
	if len(opts.SelectFields) > 0 {
		db = db.Select(opts.SelectFields)
	}
	// 设置要忽略的字段
	if len(opts.OmitFields) > 0 {
		db = db.Omit(opts.OmitFields...)
	}
	// 设置排序规则
	if len(opts.OrderFields) > 0 {
		db = db.Order(opts.OrderFields.ToSQL())
	}
	return db
}

// WrapPageQuery 执行分页查询
// 支持仅统计总数、限制返回数量和标准分页查询
func WrapPageQuery(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out interface{}) (*PaginationResult, error) {
	// 如果只需要统计总数
	if pp.OnlyCount {
		var count int64
		err := db.Count(&count).Error
		if err != nil {
			return nil, err
		}
		return &PaginationResult{Total: count}, nil
	} else if !pp.Pagination { // 如果不需要分页,但可能需要限制返回数量
		pageSize := pp.PageSize
		if pageSize > 0 {
			db = db.Limit(pageSize)
		}

		db = wrapQueryOptions(db, opts)
		err := db.Find(out).Error
		return nil, err
	}

	// 执行标准分页查询
	total, err := FindPage(ctx, db, pp, opts, out)
	if err != nil {
		return nil, err
	}

	return &PaginationResult{
		Total:    total,
		Current:  pp.Current,
		PageSize: pp.PageSize,
	}, nil
}

// FindPage 实现分页查询的核心逻辑
// 返回总记录数和查询结果
func FindPage(ctx context.Context, db *gorm.DB, pp PaginationParam, opts QueryOptions, out interface{}) (int64, error) {
	db = db.WithContext(ctx)
	// 首先统计总记录数
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	} else if count == 0 {
		return count, nil
	}

	// 计算分页偏移量并设置限制
	current, pageSize := pp.Current, pp.PageSize
	if current > 0 && pageSize > 0 {
		db = db.Offset((current - 1) * pageSize).Limit(pageSize)
	} else if pageSize > 0 {
		db = db.Limit(pageSize)
	}

	// 应用查询选项并执行查询
	db = wrapQueryOptions(db, opts)
	err = db.Find(out).Error
	return count, err
}

// FindOne 查询单条记录
// 返回是否找到记录和可能的错误
func FindOne(ctx context.Context, db *gorm.DB, opts QueryOptions, out interface{}) (bool, error) {
	db = db.WithContext(ctx)
	db = wrapQueryOptions(db, opts)
	result := db.First(out)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // 未找到记录,返回false但不返回错误
		}
		return false, err
	}
	return true, nil
}

// Exists 检查记录是否存在
// 返回是否存在和可能的错误
func Exists(ctx context.Context, db *gorm.DB) (bool, error) {
	db = db.WithContext(ctx)
	var count int64
	result := db.Count(&count)
	if err := result.Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
