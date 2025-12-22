package oss

import (
	"blog_api/src/model"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"

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
	objectKey := generateFilePath(s.config.Prefix, header.Filename)
	err = bucket.PutObject(objectKey, file, oss.ContentType(header.Header.Get("Content-Type")))
	if err != nil {
		return "", fmt.Errorf("failed to upload file to oss: %w", err)
	}
	if s.config.CustomDomain != "" {
		customDomain := strings.TrimRight(s.config.CustomDomain, "/")
		return fmt.Sprintf("%s/%s", customDomain, objectKey), nil
	}

	// 否则，返回标准的 OSS 访问 URL
	encodedObjectKey := url.PathEscape(objectKey)
	encodedObjectKey = strings.ReplaceAll(encodedObjectKey, "%2F", "/")
	return fmt.Sprintf("https://%s.%s/%s", s.config.Bucket, s.config.Endpoint, encodedObjectKey), nil
}
