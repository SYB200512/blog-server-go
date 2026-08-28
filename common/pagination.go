package common

// PageResult 分页结果
type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// PageParams 分页参数
type PageParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Normalize 规范化分页参数（默认第 1 页，每页 10 条，上限 100）
func (p *PageParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算偏移量
func (p *PageParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}
