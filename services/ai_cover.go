package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"blog-backend/config"
)

// aiMessageContent DashScope 返回内容项（文生图时仅含 image）
type aiMessageContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

// GenerateArticleCover 调用 Qwen-Image 按文章标题/摘要生成封面，
// 下载保存到本地 covers 目录，返回可访问的相对路径（如 /api/v1/covers/cover_xxx.png）。
// 未配置 DASHSCOPE_API_KEY 时返回空串，由调用方决定是否兜底。
func GenerateArticleCover(title, summary string) (string, error) {
	cfg := config.Cfg
	if cfg == nil || cfg.AI.DashScopeAPIKey == "" {
		return "", nil
	}

	payload := map[string]any{
		"model": "qwen-image-3.0",
		"input": map[string]any{
			"messages": []map[string]any{
				{"role": "user", "content": []map[string]any{
					{"text": buildCoverPrompt(title, summary)},
				}},
			},
		},
		"parameters": map[string]any{
			"prompt_extend": true,
			"watermark":     false,
			"size":          "1472*1104",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(cfg.AI.DashScopeBase, "/") +
		"/api/v1/services/aigc/multimodal-generation/generation"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
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
	return saveCover(imageURL)
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

// saveCover 把生成的图片下载保存到本地 covers 目录，返回相对访问路径
func saveCover(srcURL string) (string, error) {
	dir := config.Cfg.AI.CoversDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("cover_%d.png", time.Now().UnixNano())
	dst := filepath.Join(dir, filename)

	resp, err := http.Get(srcURL)
	if err != nil {
		return "", fmt.Errorf("下载封面失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载封面状态码 %d", resp.StatusCode)
	}
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return "/api/v1/covers/" + filename, nil
}