package logging

import (
	"time"

	jsoniter "github.com/json-iterator/go" // 使用 json-iterator 进行高性能的 JSON 处理
	"github.com/rs/xid"                    // 用于生成唯一 ID
	"gorm.io/gorm"                         // 使用 GORM ORM 框架
)

// Logger 定义了日志记录的数据结构
type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`  // 唯一标识符，作为主键
	Level     string    `gorm:"size:20;index;" json:"level"`    // 日志级别（如：info, error, debug 等）
	TraceID   string    `gorm:"size:64;index;" json:"trace_id"` // 追踪 ID，用于分布式追踪
	UserID    string    `gorm:"size:20;index;" json:"user_id"`  // 用户 ID，记录相关用户
	Tag       string    `gorm:"size:32;index;" json:"tag"`      // 日志标签，用于分类
	Message   string    `gorm:"size:1024;" json:"message"`      // 日志消息内容
	Stack     string    `gorm:"type:text;" json:"stack"`        // 错误堆栈信息
	Data      string    `gorm:"type:text;" json:"data"`         // 额外的日志数据，JSON 格式
	CreatedAt time.Time `gorm:"index;" json:"created_at"`       // 日志创建时间
}

// NewGormHook 创建一个新的 GORM 日志钩子
// 参数 db 为初始化好的 GORM 数据库连接
// 会自动创建日志表，如果失败则 panic
func NewGormHook(db *gorm.DB) *GormHook {
	err := db.AutoMigrate(new(Logger))
	if err != nil {
		panic(err)
	}

	return &GormHook{
		db: db,
	}
}

// GormHook 实现了日志钩子接口
type GormHook struct {
	db *gorm.DB // 保存数据库连接
}

// Exec 执行日志写入操作
// 参数：
//   - extra: 额外的键值对数据
//   - b: JSON 格式的日志数据字节数组
//
// 返回可能的错误
func (h *GormHook) Exec(extra map[string]string, b []byte) error {
	// 创建新的日志记录，使用 xid 生成唯一 ID
	msg := &Logger{
		ID: xid.New().String(),
	}

	// 解析传入的 JSON 数据
	data := make(map[string]interface{})
	err := jsoniter.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	// 从 data 中提取各个字段的值并设置到 Logger 结构体中
	if v, ok := data["ts"]; ok {
		msg.CreatedAt = time.UnixMilli(int64(v.(float64))) // 转换时间戳
		delete(data, "ts")
	}
	if v, ok := data["msg"]; ok {
		msg.Message = v.(string) // 设置日志消息
		delete(data, "msg")
	}
	if v, ok := data["tag"]; ok {
		msg.Tag = v.(string) // 设置标签
		delete(data, "tag")
	}
	if v, ok := data["trace_id"]; ok {
		msg.TraceID = v.(string) // 设置追踪 ID
		delete(data, "trace_id")
	}
	if v, ok := data["user_id"]; ok {
		msg.UserID = v.(string) // 设置用户 ID
		delete(data, "user_id")
	}
	if v, ok := data["level"]; ok {
		msg.Level = v.(string) // 设置日志级别
		delete(data, "level")
	}
	if v, ok := data["stack"]; ok {
		msg.Stack = v.(string) // 设置错误堆栈
		delete(data, "stack")
	}
	delete(data, "caller") // 删除调用者信息

	// 将额外的数据合并到 data 中
	for k, v := range extra {
		data[k] = v
	}

	// 如果还有其他数据，将其序列化后存储到 Data 字段
	if len(data) > 0 {
		buf, _ := jsoniter.Marshal(data)
		msg.Data = string(buf)
	}

	// 将日志记录保存到数据库
	return h.db.Create(msg).Error
}

// Close 关闭数据库连接
func (h *GormHook) Close() error {
	db, err := h.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
