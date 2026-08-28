package dto

// CategoryCreateRequest 创建分类请求
type CategoryCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CategoryUpdateRequest 更新分类请求
type CategoryUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TagCreateRequest 创建标签请求
type TagCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// CommentCreateRequest 发表评论请求
type CommentCreateRequest struct {
	Content string `json:"content" binding:"required"`
	Email   string `json:"email"`
	ParentID uint64 `json:"parent_id"`
}

// CommentQuery 评论查询参数
type CommentQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
	Status   *int8 `form:"status"` // 后台筛选用
}

// ReviewCommentRequest 审核评论请求
type ReviewCommentRequest struct {
	Status int8 `json:"status"` // 0 待审核，1 已通过
}
