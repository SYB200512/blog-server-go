package dao

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"blog-backend/model/errs"
)

// wrapErr 把存储层错误翻译成领域错误。
// 只输出错误类别和安全的通用文案，底层 SQL 原因留在包装链里，供日志排查，
// 不会被 handler 回显给客户端。
func wrapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return errs.Wrap(errs.ErrNotFound, "资源不存在", fmt.Errorf("%s: %w", op, err))
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return errs.Wrap(errs.ErrDuplicated, "资源已存在", fmt.Errorf("%s: %w", op, err))
	default:
		return errs.Wrap(errs.ErrStorage, "存储访问失败", fmt.Errorf("%s: %w", op, err))
	}
}
