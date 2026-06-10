package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type UploadService struct {
	uploadDir string
	baseURL   string
}

func NewUploadService(uploadDir, baseURL string) *UploadService {
	os.MkdirAll(uploadDir, 0755)
	return &UploadService{uploadDir: uploadDir, baseURL: baseURL}
}

func (s *UploadService) Save(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(fh.Filename)
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
	dstPath := filepath.Join(s.uploadDir, name)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return s.baseURL + "/" + name, nil
}
