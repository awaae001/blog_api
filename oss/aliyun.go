package oss

import (
	"blog_api/src/model"
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

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
