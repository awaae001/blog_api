package service

import (
	"blog_api/src/config"
	"blog_api/src/model"
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// OSSService 定义了对象存储服务的通用接口
type OSSService interface {
	// UploadFile 将文件上传到 OSS，并返回文件的 URL
	// file: 文件内容
	// header: 包含文件名等元数据的文件头
	UploadFile(file multipart.File, header *multipart.FileHeader) (string, error)
}

// NewOSSService 是一个工厂函数，根据配置创建并返回一个具体的 OSSService 实例
func NewOSSService() (OSSService, error) {
	cfg := config.GetConfig()
	if !cfg.OSS.Enable {
		return nil, fmt.Errorf("OSS service is not enabled in the configuration")
	}

	switch cfg.OSS.Provider {
	case "aliyun":
		return NewAliyunOSSService(&cfg.OSS)
	// case "tencent":
	// 	// 可以在此添加腾讯云 COS 的实现
	// 	return nil, fmt.Errorf("tencent COS provider is not yet implemented")
	case "s3":
		return NewS3OSSService(&cfg.OSS)
	default:
		return nil, fmt.Errorf("unsupported OSS provider: %s", cfg.OSS.Provider)
	}
}

// generateFilePath 生成在 OSS 中存储的文件路径
// 使用 prefix 和原始文件名，并可以添加时间戳或 UUID 以避免冲突
func generateFilePath(prefix, originalFilename string) string {
	// 为了避免文件名冲突，我们结合了时间戳和原始文件名
	timestamp := time.Now().Unix()
	uniqueFilename := fmt.Sprintf("%d-%s", timestamp, originalFilename)

	if prefix == "" {
		return uniqueFilename
	}
	return fmt.Sprintf("%s/%s", prefix, uniqueFilename)
}

// AliyunOSSService 实现了 OSSService 接口，用于阿里云 OSS
type AliyunOSSService struct {
	client *oss.Client
	config *model.OSSConfig
}

// NewAliyunOSSService 创建一个新的 AliyunOSSService 实例
func NewAliyunOSSService(cfg *model.OSSConfig) (OSSService, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create aliyun oss client: %w", err)
	}
	return &AliyunOSSService{
		client: client,
		config: cfg,
	}, nil
}

// UploadFile 实现了文件上传到阿里云 OSS 的逻辑
func (s *AliyunOSSService) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	bucket, err := s.client.Bucket(s.config.Bucket)
	if err != nil {
		return "", fmt.Errorf("failed to get oss bucket: %w", err)
	}

	// 生成在 OSS 中的存储路径
	objectKey := generateFilePath(s.config.Prefix, header.Filename)

	// 执行上传，并设置正确的 Content-Type
	// 这能确保图片等文件在浏览器中被正确预览，而不是被当做文件下载
	err = bucket.PutObject(objectKey, file, oss.ContentType(header.Header.Get("Content-Type")))
	if err != nil {
		return "", fmt.Errorf("failed to upload file to oss: %w", err)
	}

	// 根据配置返回访问 URL
	if s.config.CustomDomain != "" {
		return fmt.Sprintf("%s/%s", s.config.CustomDomain, objectKey), nil
	}
	// 否则，返回标准的 OSS 访问 URL
	encodedObjectKey := url.PathEscape(objectKey)
	return fmt.Sprintf("https://%s.%s/%s", s.config.Bucket, s.config.Endpoint, encodedObjectKey), nil
}

// S3OSSService 实现了 OSSService 接口，用于与 S3 兼容的对象存储
type S3OSSService struct {
	uploader *manager.Uploader
	config   *model.OSSConfig
}

// NewS3OSSService 创建一个新的 S3OSSService 实例
func NewS3OSSService(cfg *model.OSSConfig) (OSSService, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	uploader := manager.NewUploader(s3Client)

	return &S3OSSService{
		uploader: uploader,
		config:   cfg,
	}, nil
}

// UploadFile 实现了文件上传到 S3 的逻辑
func (s *S3OSSService) UploadFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 生成在 OSS 中的存储路径
	objectKey := generateFilePath(s.config.Prefix, header.Filename)

	// 执行上传
	_, err := s.uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to s3: %w", err)
	}

	// 根据配置返回访问 URL
	if s.config.CustomDomain != "" {
		return fmt.Sprintf("%s/%s", s.config.CustomDomain, objectKey), nil
	}

	// 否则，返回标准的 S3 访问 URL
	encodedObjectKey := url.PathEscape(objectKey)
	if s.config.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.config.Endpoint, s.config.Bucket, encodedObjectKey), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.Bucket, s.config.Region, encodedObjectKey), nil
}
