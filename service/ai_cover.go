package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/appTypes"
	"blog-backend/model/database"
	"blog-backend/utils/upload"
)

// aiMessageContent DashScope 返回内容项（文生图时仅含 image）
type aiMessageContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

// generateArticleCover 调用 Z-Image-Turbo 按文章标题/摘要生成封面，
// 下载保存到本地 covers 目录，返回可访问的相对路径（如 /api/v1/covers/cover_xxx.png）。
// ctx 用于取消请求与下载：调用方（异步任务）自己控制超时预算。
// 未配置 DASHSCOPE_API_KEY 时返回空串，由调用方决定是否兜底。
func generateArticleCover(ctx context.Context, title, summary string) (string, error) {
	cfg := global.Config
	if cfg == nil || cfg.AI.DashScopeAPIKey == "" {
		return "", nil
	}

	payload := map[string]any{
		"model": "z-image-turbo",
		"input": map[string]any{
			"messages": []map[string]any{
				{"role": "user", "content": []map[string]any{
					{"text": buildCoverPrompt(title, summary)},
				}},
			},
		},
		"parameters": map[string]any{
			"prompt_extend": true,
			"size":          "1472*1104",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(cfg.AI.DashScopeBase, "/") +
		"/api/v1/services/aigc/multimodal-generation/generation"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AI.DashScopeAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("文生图请求失败: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("文生图返回 %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Output struct {
			Choices []struct {
				Message struct {
					Content []aiMessageContent `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析文生图响应失败: %w", err)
	}

	var imageURL string
	for _, choice := range result.Output.Choices {
		for _, content := range choice.Message.Content {
			if content.Image != "" {
				imageURL = content.Image
				break
			}
		}
		if imageURL != "" {
			break
		}
	}
	if imageURL == "" {
		return "", fmt.Errorf("文生图结果中未找到图片 URL")
	}
	return saveCover(ctx, imageURL)
}

// buildCoverPrompt 根据标题与摘要拼接文生图提示词
func buildCoverPrompt(title, summary string) string {
	subject := strings.TrimSpace(summary)
	if subject == "" {
		subject = title
	}
	return fmt.Sprintf(
		"为技术博客文章生成一张现代极简风格、横版构图的封面插图，主题贴合标题和内容概要，画面中不要出现任何文字或水印。标题：%s。内容概要：%s",
		title, subject)
}

// saveCover 把生成的图片保存到当前配置的存储（本地目录或 COS），
// 同时写入图片台账（分类：封面），返回可访问的 URL
func saveCover(ctx context.Context, srcURL string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return "", fmt.Errorf("构造封面下载请求失败: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载封面失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载封面状态码 %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取封面数据失败: %w", err)
	}

	filename := fmt.Sprintf("cover_%d.png", time.Now().UnixNano())
	url, key, err := upload.NewOss().UploadBytes(ctx, data, filename)
	if err != nil {
		return "", fmt.Errorf("保存封面失败: %w", err)
	}

	if err := dao.CreateImage(ctx, &database.Image{
		Name:     key,
		URL:      url,
		Category: appTypes.Cover,
		Storage:  upload.CurrentStorage(),
		Size:     int64(len(data)),
	}); err != nil {
		global.Log.Warn("写入封面图片台账失败", zap.String("url", url), zap.Error(err))
	}
	return url, nil
}
