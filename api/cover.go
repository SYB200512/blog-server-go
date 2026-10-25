package api

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"blog-backend/global"
	"blog-backend/model/response"
)

// CoverApi 封面资源接口
type CoverApi struct{}

// GetCover 前台：返回 AI 生成的封面图片文件
// @Summary 封面图片
// @Tags 前台-资源
// @Produce image/png
// @Param filename path string true "封面文件名"
// @Success 200 {file} binary "图片文件"
// @Failure 404 {object} response.Response "文件不存在"
// @Router /api/v1/covers/{filename} [get]
// @Id covers_get
func (a *CoverApi) GetCover(c *gin.Context) {
	dir := global.Config.AI.CoversDir
	// 仅取文件名，防止路径穿越
	filename := filepath.Base(c.Param("filename"))
	fullPath := filepath.Join(dir, filename)

	// 校验目标文件确实位于 covers 目录内
	resolved, err := filepath.Abs(fullPath)
	if err != nil {
		global.Log.Warn("解析封面文件路径失败", zap.String("filename", filename), zap.Error(err))
		response.NotFound(c, "文件不存在")
		return
	}
	c.File(resolved)
}
