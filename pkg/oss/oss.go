// Package oss 提供对象存储服务的抽象接口和基础实现
package oss

import (
	"context"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/xid" // 用于生成唯一ID
)

var (
	// Ins 全局的OSS客户端实例
	Ins IClient
	// once 确保全局实例只被初始化一次的同步锁
	once sync.Once
)

// SetGlobal 设置全局OSS客户端实例
// h 是一个返回 IClient 接口实现的函数
func SetGlobal(h func() IClient) {
	once.Do(func() {
		Ins = h()
	})
}

// IClient 定义了OSS客户端需要实现的接口方法
type IClient interface {
	// PutObject 上传对象到OSS
	// bucketName: 存储桶名称
	// objectName: 对象名称
	// reader: 对象数据读取器
	// objectSize: 对象大小
	// options: 可选的上传选项
	PutObject(ctx context.Context, bucketName, objectName string, reader io.ReadSeeker, objectSize int64, options ...PutObjectOptions) (*PutObjectResult, error)

	// GetObject 从OSS获取对象
	GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)

	// RemoveObject 从OSS删除对象
	RemoveObject(ctx context.Context, bucketName, objectName string) error

	// RemoveObjectByURL 通过URL删除对象
	RemoveObjectByURL(ctx context.Context, urlStr string) error

	// StatObject 获取对象的元数据信息
	StatObject(ctx context.Context, bucketName, objectName string) (*ObjectStat, error)

	// StatObjectByURL 通过URL获取对象的元数据信息
	StatObjectByURL(ctx context.Context, urlStr string) (*ObjectStat, error)
}

// PutObjectOptions 定义上传对象时的可选参数
type PutObjectOptions struct {
	ContentType  string            // 内容类型
	UserMetadata map[string]string // 用户自定义元数据
}

// PutObjectResult 定义上传对象后的返回结果
type PutObjectResult struct {
	URL  string `json:"url,omitempty"`   // 对象访问URL
	Key  string `json:"key,omitempty"`   // 对象键名
	ETag string `json:"e_tag,omitempty"` // 对象的ETag值
	Size int64  `json:"size,omitempty"`  // 对象大小
}

// ObjectStat 定义对象的元数据信息
type ObjectStat struct {
	Key          string            // 对象键名
	ETag         string            // 对象的ETag值
	LastModified time.Time         // 最后修改时间
	Size         int64             // 对象大小
	ContentType  string            // 内容类型
	UserMetadata map[string]string // 用户自定义元数据
}

// GetName 获取对象名称
// 如果在用户元数据中定义了name，则返回该值
// 否则返回对象键名中的基础名称部分
func (a *ObjectStat) GetName() string {
	if name, ok := a.UserMetadata["name"]; ok {
		return name
	}
	return filepath.Base(a.Key)
}

// formatObjectName 格式化对象名称
// prefix: 对象名称的前缀
// objectName: 原始对象名称
// 如果objectName为空，则生成一个唯一ID作为对象名称
// 确保对象名称格式正确（去除开头的'/'，添加前缀等）
func formatObjectName(prefix, objectName string) string {
	if objectName == "" {
		objectName = xid.New().String()
	}
	if objectName[0] == '/' {
		objectName = objectName[1:]
	}
	if prefix != "" {
		objectName = prefix + "/" + objectName
	}
	return objectName
}
