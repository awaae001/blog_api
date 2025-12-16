package service

import (
	"blog_api/src/model"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// ResourceService 提供了处理资源（如文件上传）的服务。
type ResourceService struct {
	config *model.Config
}

// NewResourceService 创建一个新的 ResourceService 实例。
func NewResourceService(cfg *model.Config) *ResourceService {
	return &ResourceService{config: cfg}
}

// SaveFile 保存上传的文件。
// 它会检查文件扩展名是否在白名单内，并清理目标路径以防止路径遍历。
// overwrite 参数决定如果文件已存在，是覆盖它还是生成一个新名字。
func (s *ResourceService) SaveFile(file *multipart.FileHeader, subPath string, overwrite bool) (string, error) {
	// 检查文件扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	if !s.isExtensionAllowed(ext) {
		return "", fmt.Errorf("文件类型 '%s' 不被允许", ext)
	}

	// 清理并构建保存路径，防止路径遍历攻击
	cleanSubPath := filepath.Clean(subPath)
	if strings.HasPrefix(cleanSubPath, "..") {
		return "", fmt.Errorf("无效的路径")
	}

	basePath := s.config.Data.Resource.Path
	if basePath == "" {
		basePath = "data/" // 默认路径
	}
	saveDir := filepath.Join(basePath, cleanSubPath)

	// 创建目录（如果不存在）
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 根据 overwrite 标志决定文件名
	var finalFilename string
	if overwrite {
		finalFilename = file.Filename
	} else {
		finalFilename = s.findUniqueFilename(saveDir, file.Filename)
	}
	filePath := filepath.Join(saveDir, finalFilename)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	return filePath, nil
}

// isExtensionAllowed 检查文件扩展名是否在白名单中。
func (s *ResourceService) isExtensionAllowed(ext string) bool {
	for _, allowedExt := range s.config.Safe.AllowExtension {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// findUniqueFilename 检查文件名是否重复，如果重复则添加后缀 (1), (2)...
func (s *ResourceService) findUniqueFilename(dir, filename string) string {
	filePath := filepath.Join(dir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filename // 文件名不重复，直接返回
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	counter := 1
	for {
		// 生成新的文件名，例如: "image(1).png"
		newFilename := fmt.Sprintf("%s(%d)%s", baseName, counter, ext)
		newFilePath := filepath.Join(dir, newFilename)
		if _, err := os.Stat(newFilePath); os.IsNotExist(err) {
			return newFilename // 找到一个不重复的文件名
		}
		counter++
	}
}

// DeleteFile 删除指定路径的文件。
// 在删除前会进行严格的安全检查，以防止删除受保护的文件。
func (s *ResourceService) DeleteFile(filePath string) error {
	// 1. 清理路径，防止路径遍历
	cleanPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 2. 检查路径是否在受保护的目录内
	for _, protectedPath := range s.config.Safe.ExcludePaths {
		absProtectedPath, err := filepath.Abs(protectedPath)
		if err != nil {
			// 在配置加载时就应保证路径有效，但这里还是做个检查
			return fmt.Errorf("获取受保护路径的绝对路径失败: '%s'", protectedPath)
		}
		if strings.HasPrefix(cleanPath, absProtectedPath) {
			return fmt.Errorf("禁止删除受保护的路径下的文件: '%s'", filePath)
		}
	}

	// 3. 检查文件是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: '%s'", filePath)
	}

	// 4. 执行删除
	if err := os.Remove(cleanPath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}
