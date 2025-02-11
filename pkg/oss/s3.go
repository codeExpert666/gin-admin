// Package oss 提供对象存储服务的实现
package oss

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3ClientConfig 定义了 S3 客户端的配置参数
type S3ClientConfig struct {
	Domain          string // 访问域名
	Region          string // AWS 区域
	AccessKeyID     string // 访问密钥 ID
	SecretAccessKey string // 访问密钥
	BucketName      string // 存储桶名称
	Prefix          string // 对象键前缀
}

// 确保 S3Client 实现了 IClient 接口
var _ IClient = (*S3Client)(nil)

// S3Client 实现了 AWS S3 存储服务的客户端
type S3Client struct {
	config  S3ClientConfig   // 客户端配置
	session *session.Session // AWS 会话
	client  *s3.S3           // S3 服务客户端
}

// NewS3Client 创建一个新的 S3 客户端实例
func NewS3Client(config S3ClientConfig) (*S3Client, error) {
	// 创建 AWS 配置
	awsConfig := aws.NewConfig()
	awsConfig.WithRegion(config.Region)
	awsConfig.WithCredentials(credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""))

	// 创建新的 AWS 会话
	session, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return &S3Client{
		config:  config,
		session: session,
		client:  s3.New(session),
	}, nil
}

// PutObject 上传对象到 S3 存储
func (c *S3Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.ReadSeeker, objectSize int64, options ...PutObjectOptions) (*PutObjectResult, error) {
	// 如果未指定存储桶，使用默认配置中的存储桶
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

	// 构建上传请求参数
	input := &s3.PutObjectInput{
		Bucket:             aws.String(bucketName),
		Key:                aws.String(objectName),
		Body:               reader,
		ContentType:        aws.String(opt.ContentType),
		ContentDisposition: aws.String("inline"),
		ACL:                aws.String("public-read"), // 设置为公开可读
	}

	// 添加用户自定义元数据
	if len(opt.UserMetadata) > 0 {
		input.Metadata = make(map[string]*string)
		for k, v := range opt.UserMetadata {
			input.Metadata[k] = aws.String(v)
		}
	}

	// 执行上传操作
	output, err := c.client.PutObject(input)
	if err != nil {
		return nil, err
	}

	// 返回上传结果
	return &PutObjectResult{
		URL:  c.config.Domain + "/" + objectName,
		Key:  *input.Key,
		ETag: *output.ETag,
		Size: objectSize,
	}, nil
}

// GetObject 从 S3 存储获取对象
func (c *S3Client) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	output, err := c.client.GetObject(input)
	if err != nil {
		return nil, err
	}

	return output.Body, nil
}

// RemoveObject 从 S3 存储删除指定对象
func (c *S3Client) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	_, err := c.client.DeleteObject(input)
	return err
}

// RemoveObjectByURL 通过 URL 删除对象
// 会验证 URL 是否属于当前配置的域名
func (c *S3Client) RemoveObjectByURL(ctx context.Context, urlStr string) error {
	prefix := c.config.Domain + "/"
	if !strings.HasPrefix(urlStr, prefix) {
		return nil
	}

	objectName := strings.TrimPrefix(urlStr, prefix)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(objectName),
	}

	_, err := c.client.DeleteObject(input)
	return err
}

// StatObjectByURL 通过 URL 获取对象状态信息
func (c *S3Client) StatObjectByURL(ctx context.Context, urlStr string) (*ObjectStat, error) {
	prefix := c.config.Domain + "/"
	if !strings.HasPrefix(urlStr, prefix) {
		return nil, nil
	}

	objectName := strings.TrimPrefix(urlStr, prefix)
	return c.StatObject(ctx, c.config.BucketName, objectName)
}

// StatObject 获取对象的详细信息，包括大小、类型、最后修改时间等
func (c *S3Client) StatObject(ctx context.Context, bucketName, objectName string) (*ObjectStat, error) {
	if bucketName == "" {
		bucketName = c.config.BucketName
	}

	objectName = formatObjectName(c.config.Prefix, objectName)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}

	// 获取对象元数据
	output, err := c.client.HeadObject(input)
	if err != nil {
		return nil, err
	}

	// 处理用户自定义元数据
	var metadata map[string]string
	if output.Metadata != nil {
		metadata = make(map[string]string)
		for k, v := range output.Metadata {
			metadata[k] = *v
		}
	}

	// 返回对象状态信息
	return &ObjectStat{
		Key:          objectName,
		ETag:         *output.ETag,
		LastModified: *output.LastModified,
		Size:         *output.ContentLength,
		ContentType:  *output.ContentType,
		UserMetadata: metadata,
	}, nil
}
