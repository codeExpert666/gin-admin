// package util 提供了一些通用的工具函数和数据结构
package util

import "github.com/LyricTian/gin-admin/v10/pkg/errors"

// 定义一些常用的键名常量
const (
	// ReqBodyKey 请求体的键名
	ReqBodyKey = "req-body"
	// ResBodyKey 响应体的键名
	ResBodyKey = "res-body"
	// TreePathDelimiter 树形结构路径的分隔符
	TreePathDelimiter = "."
)

// ResponseResult 定义了API响应的标准格式
type ResponseResult struct {
	Success bool          `json:"success"`         // 表示请求是否成功
	Data    interface{}   `json:"data,omitempty"`  // 响应的数据内容,可以是任意类型
	Total   int64         `json:"total,omitempty"` // 数据总数,通常用于分页
	Error   *errors.Error `json:"error,omitempty"` // 错误信息,当Success为false时使用
}

// PaginationResult 定义了分页查询的结果格式
type PaginationResult struct {
	Total    int64 `json:"total"`    // 数据总数
	Current  int   `json:"current"`  // 当前页码
	PageSize int   `json:"pageSize"` // 每页数据量
}

// PaginationParam 定义了分页查询的参数
type PaginationParam struct {
	Pagination bool `form:"-"`                          // 是否使用分页
	OnlyCount  bool `form:"-"`                          // 是否只返回总数
	Current    int  `form:"current"`                    // 当前页码
	PageSize   int  `form:"pageSize" binding:"max=100"` // 每页数据量,最大100
}

// QueryOptions 定义了数据查询的选项
type QueryOptions struct {
	SelectFields []string      // 要选择的字段列表
	OmitFields   []string      // 要忽略的字段列表
	OrderFields  OrderByParams // 排序参数
}

// Direction 定义了排序的方向类型
type Direction string

// 定义排序方向的常量
const (
	ASC  Direction = "ASC"  // 升序
	DESC Direction = "DESC" // 降序
)

// OrderByParam 定义了单个排序参数的结构
type OrderByParam struct {
	Field     string    // 排序字段名
	Direction Direction // 排序方向
}

// OrderByParams 是OrderByParam的切片类型,用于支持多字段排序
type OrderByParams []OrderByParam

// ToSQL 将排序参数转换为SQL排序语句
func (a OrderByParams) ToSQL() string {
	if len(a) == 0 {
		return ""
	}

	var sql string
	for _, v := range a {
		sql += v.Field + " " + string(v.Direction) + ","
	}
	return sql[:len(sql)-1] // 移除最后一个逗号
}
