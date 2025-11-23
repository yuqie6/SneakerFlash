package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadService 负责图片落盘校验
type UploadService struct {
	uploadDir string
	maxSize   int64
}

func NewUploadService(uploadDir string) *UploadService {
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	uploadDir = filepath.Clean(uploadDir)
	// 默认 5MB
	return &UploadService{uploadDir: uploadDir, maxSize: 5 << 20}
}

// SaveImage 保存图片并返回可直接访问的相对路径
func (s *UploadService) SaveImage(file *multipart.FileHeader) (string, error) {
	if file.Size > s.maxSize {
		return "", fmt.Errorf("文件过大，最大支持 %.1fMB", float64(s.maxSize)/(1<<20))
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return "", fmt.Errorf("缺少文件扩展名")
	}
	allowed := map[string]struct{}{".jpg": {}, ".jpeg": {}, ".png": {}, ".gif": {}, ".webp": {}}
	if _, ok := allowed[ext]; !ok {
		return "", fmt.Errorf("仅支持 jpg/jpeg/png/gif/webp")
	}

	// 基础的内容探测
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("文件打开失败")
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("仅支持图片上传")
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("创建上传目录失败")
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(s.uploadDir, filename)

	// 重新打开文件进行落盘
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("文件打开失败")
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("创建文件失败")
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, src); err != nil {
		return "", fmt.Errorf("保存文件失败")
	}

	publicPath := filepath.ToSlash(filepath.Join("/uploads", filename))
	return publicPath, nil
}
