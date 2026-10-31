package upload

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"blog-backend/global"
	"blog-backend/model/appTypes"
)

// WhiteImageList 允许上传的图片后缀白名单
var WhiteImageList = map[string]struct{}{
	".jpg":  {},
	".png":  {},
	".jpeg": {},
	".ico":  {},
	".tiff": {},
	".gif":  {},
	".svg":  {},
	".webp": {},
}

// OSS 对象存储抽象：本地目录与腾讯云 COS 实现同一套接口。
// 所有方法都接收 ctx：请求取消或超时时，底层上传/删除应立即中断。
type OSS interface {
	// UploadImage 上传 multipart 文件，返回访问 URL 与对象键
	UploadImage(ctx context.Context, file *multipart.FileHeader) (url, key string, err error)
	// UploadBytes 上传内存中的字节（用于 AI 生成图直接落盘/上云）
	UploadBytes(ctx context.Context, data []byte, filename string) (url, key string, err error)
	// Delete 按对象键删除文件
	Delete(ctx context.Context, key string) error
}

// NewOss 按当前配置的 storage.type 返回对应实现（上传时使用）
func NewOss() OSS {
	switch strings.ToLower(global.Config.Storage.Type) {
	case "cos":
		return &Cos{}
	default:
		return &Local{}
	}
}

// NewOssWithStorage 按图片记录中的存储方式返回实现（删除时使用）
func NewOssWithStorage(storage appTypes.Storage) OSS {
	switch storage {
	case appTypes.Cos:
		return &Cos{}
	default:
		return &Local{}
	}
}

// CurrentStorage 当前配置对应的存储枚举
func CurrentStorage() appTypes.Storage {
	return appTypes.ToStorage(global.Config.Storage.Type)
}

// checkImage 校验图片大小与后缀
func checkImage(file *multipart.FileHeader) error {
	maxMB := global.Config.Storage.MaxSize
	if maxMB > 0 {
		sizeMB := float64(file.Size) / (1024 * 1024)
		if sizeMB > float64(maxMB) {
			return fmt.Errorf("图片大小超过限制：当前 %.2f MB，上限 %d MB", sizeMB, maxMB)
		}
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := WhiteImageList[ext]; !ok {
		return errors.New("不支持的文件类型，仅允许上传图片")
	}
	return nil
}

// buildObjectName 生成唯一文件名：原文件名 MD5 前 8 位 + 时间戳 + 后缀
func buildObjectName(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	sum := md5.Sum([]byte(name))
	return hex.EncodeToString(sum[:])[:8] + "-" + time.Now().Format("20060102150405") + ext
}
