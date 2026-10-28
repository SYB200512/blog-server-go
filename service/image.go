package service

import (
	"mime/multipart"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"blog-backend/dao"
	"blog-backend/global"
	"blog-backend/model/appTypes"
	"blog-backend/model/database"
	"blog-backend/model/request"
	"blog-backend/model/response"
	"blog-backend/utils/upload"
)

// ImageService 图片业务
type ImageService struct{}

// UploadImage 上传图片到当前配置的存储，并写入图片台账
func (s *ImageService) UploadImage(file *multipart.FileHeader, category appTypes.ImageCategory) (*response.ImageUploadResult, error) {
	oss := upload.NewOss()
	url, key, err := oss.UploadImage(file)
	if err != nil {
		global.Log.Error("上传图片失败", zap.String("filename", file.Filename), zap.Error(err))
		return nil, err
	}

	image := &database.Image{
		Name:     key,
		URL:      url,
		Category: category,
		Storage:  upload.CurrentStorage(),
		Size:     file.Size,
	}
	if err := dao.CreateImage(image); err != nil {
		global.Log.Error("写入图片台账失败", zap.String("url", url), zap.Error(err))
		// 台账写失败则回收刚上传的文件，避免产生孤儿图
		_ = oss.Delete(key)
		return nil, err
	}
	return &response.ImageUploadResult{
		ID:      image.ID,
		URL:     image.URL,
		Name:    image.Name,
		Storage: image.Storage,
	}, nil
}

// ListImages 图片列表
func (s *ImageService) ListImages(q *request.ImageListQuery) (*response.PageResult, error) {
	p := &request.PageParams{Page: q.Page, PageSize: q.PageSize}
	p.Normalize()

	var category *appTypes.ImageCategory
	if q.Category != nil && *q.Category != "" {
		c := appTypes.ToImageCategory(*q.Category)
		category = &c
	}

	images, total, err := dao.ListImages(p, category, q.Keyword)
	if err != nil {
		global.Log.Error("查询图片列表失败", zap.Error(err))
		return nil, err
	}
	return &response.PageResult{List: images, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// DeleteImages 批量删除图片（先删存储文件，再删台账记录）
func (s *ImageService) DeleteImages(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	images, err := dao.GetImagesByIDs(ids)
	if err != nil {
		global.Log.Error("查询待删除图片失败", zap.Error(err))
		return err
	}
	for _, image := range images {
		if err := upload.NewOssWithStorage(image.Storage).Delete(image.Name); err != nil {
			global.Log.Warn("删除存储文件失败（继续删台账）",
				zap.String("key", image.Name), zap.Error(err))
		}
	}
	if err := dao.DeleteImagesByIDs(ids); err != nil {
		global.Log.Error("删除图片台账失败", zap.Error(err))
		return err
	}
	return nil
}

// UpdateCategory 批量修改图片分类
func (s *ImageService) UpdateCategory(ids []uint64, category appTypes.ImageCategory) error {
	if err := dao.UpdateImageCategoryByIDs(ids, category); err != nil {
		global.Log.Error("批量修改图片分类失败", zap.Error(err))
		return err
	}
	return nil
}

// ListPhotos 图片墙：公开读取分类为「图片墙」的图片
func (s *ImageService) ListPhotos(limit int) ([]database.Image, error) {
	if limit <= 0 || limit > 100 {
		limit = 60
	}
	return dao.ListImagesByCategory(appTypes.Photo, limit)
}

// markdownImageRegex 匹配 Markdown 图片语法 ![alt](url) 中的 URL
var markdownImageRegex = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`)

// MarkArticleImages 文章保存后，按用途回写图片分类：封面 → 封面，正文引用 → 插图
func (s *ImageService) MarkArticleImages(cover, content string) {
	if cover != "" {
		if err := dao.UpdateImageCategoryByURLs([]string{cover}, appTypes.Cover); err != nil {
			global.Log.Warn("回写封面分类失败", zap.String("cover", cover), zap.Error(err))
		}
	}

	matches := markdownImageRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(matches))
	urls := make([]string, 0, len(matches))
	for _, m := range matches {
		u := strings.TrimSpace(m[1])
		if u == "" || u == cover {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		urls = append(urls, u)
	}
	if err := dao.UpdateImageCategoryByURLs(urls, appTypes.Illustration); err != nil {
		global.Log.Warn("回写正文插图分类失败", zap.Error(err))
	}
}

// ExtractImageURLs 从正文中提取图片 URL（供其他模块复用）
func ExtractImageURLs(content string) []string {
	matches := markdownImageRegex.FindAllStringSubmatch(content, -1)
	urls := make([]string, 0, len(matches))
	for _, m := range matches {
		urls = append(urls, m[1])
	}
	return urls
}
