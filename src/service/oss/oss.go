package oss

import (
	"blog_api/src/model"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
)

// UploadInput describes a replayable object upload.
type UploadInput struct {
	Name        string
	ContentType string
	Size        int64
	Body        io.ReadSeeker
}

// OSSService defines object storage operations.
type OSSService interface {
	// Upload stores input and returns its public URL and object key.
	Upload(ctx context.Context, input UploadInput) (string, string, error)
	// DeleteFile 从 OSS 删除文件
	// objectKey: 文件在 OSS 中的 key
	DeleteFile(objectKey string) error
}

// NewOSSService creates an optional OSS client from validated startup config.
func NewOSSService(cfg *model.OSSConfig) (OSSService, error) {
	if !cfg.Enable {
		return nil, nil
	}
	if cfg.Provider == "aliyun" {
		return NewAliyunOSSService(cfg)
	}
	return NewS3OSSService(cfg)
}

// generateFilePath 生成在 OSS 中存储的文件路径
// 使用 prefix 和原始文件名，并可以添加时间戳或 UUID 以避免冲突
func generateFilePath(prefix, originalFilename string) string {
	// 确保路径分隔符为 '/', 以兼容 OSS
	ossFriendlyFilename := strings.ReplaceAll(originalFilename, "\\", "/")
	timestamp := time.Now().Unix()
	dir, file := path.Split(ossFriendlyFilename)
	uniqueFile := fmt.Sprintf("%d-%s", timestamp, file)
	fullPath := strings.Trim(path.Join(dir, uniqueFile), "/")

	cleanPrefix := strings.Trim(prefix, "/")
	if cleanPrefix == "" {
		return fullPath
	}

	return fmt.Sprintf("%s/%s", cleanPrefix, fullPath)
}
