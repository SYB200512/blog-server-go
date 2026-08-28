package dto

// ArticleQuery 文章查询参数
type ArticleQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Keyword    string `form:"keyword"`    // 标题/摘要关键词
	CategoryID uint64 `form:"category_id"` // 分类筛选
	TagID      uint64 `form:"tag_id"`     // 标签筛选
	Status     *int8  `form:"status"`     // 状态筛选（后台用）
}

// ArticleCreateRequest 创建文章请求
type ArticleCreateRequest struct {
	Title      string   `json:"title" binding:"required"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	Cover      string   `json:"cover"`
	CategoryID *uint64  `json:"category_id"`
	Status     int8     `json:"status"` // 0 草稿，1 发布
	IsTop      int8     `json:"is_top"`
	TagIDs     []uint64 `json:"tag_ids"`
}

// ArticleUpdateRequest 更新文章请求
type ArticleUpdateRequest struct {
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	Cover      string   `json:"cover"`
	CategoryID *uint64  `json:"category_id"`
	Status     *int8    `json:"status"`
	IsTop      *int8    `json:"is_top"`
	TagIDs     []uint64 `json:"tag_ids"`
}
