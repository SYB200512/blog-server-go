package upload

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"blog-backend/global"
)

// Local 本地磁盘存储：写入 covers 目录，通过 /api/v1/covers/{filename} 访问
type Local struct{}

// UploadImage 上传 multipart 文件到本地目录
func (*Local) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, string, error) {
	if err := checkImage(file); err != nil {
		return "", "", err
	}
	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return "", "", err
	}
	return (&Local{}).UploadBytes(ctx, data, file.Filename)
}

// UploadBytes 将字节写入本地 covers 目录
func (*Local) UploadBytes(_ context.Context, data []byte, filename string) (string, string, error) {
	dir := global.Config.AI.CoversDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	key := buildObjectName(filename)
	if err := os.WriteFile(filepath.Join(dir, key), data, 0o644); err != nil {
		return "", "", err
	}
	return "/api/v1/covers/" + key, key, nil
}

// Delete 删除本地文件
func (*Local) Delete(_ context.Context, key string) error {
	return os.Remove(filepath.Join(global.Config.AI.CoversDir, filepath.Base(key)))
}
