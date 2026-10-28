package request

// ImageListQuery 图片列表查询参数
type ImageListQuery struct {
	Page     int     `form:"page" json:"page"`
	PageSize int     `form:"page_size" json:"page_size"`
	Category *string `form:"category" json:"category"` // 分类中文名，空则不过滤
	Keyword  string  `form:"keyword" json:"keyword"`   // 文件名关键词
}

// ImageDeleteRequest 批量删除图片
type ImageDeleteRequest struct {
	IDs []uint64 `json:"ids" binding:"required"`
}

// ImageCategoryRequest 批量修改图片分类
type ImageCategoryRequest struct {
	IDs      []uint64 `json:"ids" binding:"required"`
	Category string   `json:"category" binding:"required"`
}
