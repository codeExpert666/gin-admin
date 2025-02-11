// Package oss 提供对象存储服务的实现
package oss

import (
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClientConfig 定义了 MinIO 客户端的配置参数
type MinioClientConfig struct {
	Domain          string // 访问域名
	Endpoint        string // MinIO 服务端点
	AccessKeyID     string // 访问密钥 ID
	SecretAccessKey string // 访问密钥密码
	BucketName      string // 存储桶名称
	Prefix          string // 对象名称前缀
}

// 确保 MinioClient 实现了 IClient 接口
var _ IClient = (*MinioClient)(nil)

// MinioClient 实现了 MinIO 客户端的具体操作
type MinioClient struct {
	config MinioClientConfig // 客户端配置
	client *minio.Client     // MinIO SDK 客户端实例
}

// NewMinioClient 创建一个新的 MinIO 客户端实例
func NewMinioClient(config MinioClientConfig) (*MinioClient, error) {
	// 创建 MinIO 客户端连接
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// 检查存储桶是否存在，不存在则创建
	if exists, err := client.BucketExists(ctx, config.BucketName); err != nil {
		return nil, err
	} else if !exists {
		if err := client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioClient{
		config: config,
		client: client,
	}, nil
}

// PutObject 上传对象到 MinIO 存储
func (c *MinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.ReadSeeker, objectSize int64, options ...PutObjectOptions) (*PutObjectResult, error) {
	// 如果未指定存储桶，使用默认配置的存储桶
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	// 处理可选参数
	var opt PutObjectOptions
	if len(options) > 0 {
		opt = options[0]
	}

	// 格式化对象名称，添加前缀
	objectName = formatObjectName(c.config.Prefix, objectName)
	// 执行上传操作
	output, err := c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType:  opt.ContentType,
		UserMetadata: opt.UserMetadata,
	})
	if err != nil {
		return nil, err
	}

	// 返回上传结果
	return &PutObjectResult{
		URL:  c.config.Domain + "/" + objectName,
		Key:  output.Key,
		ETag: output.ETag,
		Size: output.Size,
	}, nil
}

// GetObject 从 MinIO 获取对象
func (c *MinioClient) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	return c.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

// RemoveObject 删除指定的对象
func (c *MinioClient) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// RemoveObjectByURL 通过 URL 删除对象
func (c *MinioClient) RemoveObjectByURL(ctx context.Context, urlStr string) error {
	prefix := c.config.Domain + "/"
	// 检查 URL 是否属于当前域名
	if !strings.HasPrefix(urlStr, prefix) {
		return nil
	}

	// 从 URL 中提取对象名称并删除
	objectName := strings.TrimPrefix(urlStr, prefix)
	return c.RemoveObject(ctx, "", objectName)
}

// StatObjectByURL 通过 URL 获取对象状态信息
func (c *MinioClient) StatObjectByURL(ctx context.Context, urlStr string) (*ObjectStat, error) {
	prefix := c.config.Domain + "/"
	if !strings.HasPrefix(urlStr, prefix) {
		return nil, nil
	}

	objectName := strings.TrimPrefix(urlStr, prefix)
	return c.StatObject(ctx, "", objectName)
}

// StatObject 获取对象的状态信息
func (c *MinioClient) StatObject(ctx context.Context, bucketName, objectName string) (*ObjectStat, error) {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	// 获取对象的元数据信息
	info, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	// 返回对象状态信息
	return &ObjectStat{
		Key:          info.Key,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		UserMetadata: info.UserMetadata,
	}, nil
}
