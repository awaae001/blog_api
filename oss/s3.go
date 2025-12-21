package oss

import (
	"blog_api/src/model"
	"context"
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
