package upload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"

	"blog-backend/global"
)

// Cos 腾讯云对象存储
type Cos struct{}

// newCosClient 构建 COS 客户端，配置不完整时返回错误
func newCosClient() (*cos.Client, error) {
	cfg := global.Config.Storage.Cos
	if cfg.SecretID == "" || cfg.SecretKey == "" || cfg.Bucket == "" || cfg.Region == "" {
		return nil, errors.New("COS 配置不完整，请检查 secret_id/secret_key/bucket/region")
	}
	bucketURL := "https://" + cfg.Bucket + ".cos." + cfg.Region + ".myqcloud.com"
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, err
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	}), nil
}

// UploadImage 上传 multipart 文件到 COS
func (*Cos) UploadImage(file *multipart.FileHeader) (string, string, error) {
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
	return (&Cos{}).UploadBytes(data, file.Filename)
}

// UploadBytes 将字节上传到 COS，返回公网访问 URL 与对象键
func (*Cos) UploadBytes(data []byte, filename string) (string, string, error) {
	client, err := newCosClient()
	if err != nil {
		return "", "", err
	}
	cfg := global.Config.Storage.Cos
	key := strings.Trim(cfg.Prefix, "/")
	if key != "" {
		key += "/"
	}
	key += buildObjectName(filename)

	ctx := context.Background()
	if _, err := client.Object.Put(ctx, key, bytes.NewReader(data), nil); err != nil {
		return "", "", err
	}

	if domain := strings.TrimRight(cfg.Domain, "/"); domain != "" {
		if !strings.HasPrefix(domain, "http") {
			domain = "https://" + domain
		}
		return domain + "/" + key, key, nil
	}
	return "https://" + cfg.Bucket + ".cos." + cfg.Region + ".myqcloud.com/" + key, key, nil
}

// Delete 按对象键删除 COS 文件
func (*Cos) Delete(key string) error {
	client, err := newCosClient()
	if err != nil {
		return err
	}
	_, err = client.Object.Delete(context.Background(), key)
	return err
}
