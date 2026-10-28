package response

import "blog-backend/model/appTypes"

// ImageUploadResult 图片上传返回结果
type ImageUploadResult struct {
	ID      uint64           `json:"id"`
	URL     string           `json:"url"`
	Name    string           `json:"name"`
	Storage appTypes.Storage `json:"storage"`
}
