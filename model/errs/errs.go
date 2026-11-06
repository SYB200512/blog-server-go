package errs

import (
	"errors"
	"fmt"
)

// 领域错误类别，用 errors.Is(err, ErrXxx) 判断。
// 这里只表达“发生了什么类别的问题”，不包含任何存储实现细节。
var (
	ErrNotFound   = errors.New("not found")
	ErrDuplicated = errors.New("duplicated")
	ErrInvalid    = errors.New("invalid request")
	ErrStorage    = errors.New("storage failure")
)

// Error 领域错误。
// Error() 供日志使用，会带上底层原因；Message() 才是可以回给用户的安全文案。
type Error struct {
	kind  error
	msg   string
	cause error
}

// New 创建不带底层原因的领域错误
func New(kind error, msg string) *Error {
	return &Error{kind: kind, msg: msg}
}

// Wrap 创建带底层原因的领域错误，cause 应使用 %w 包装链
func Wrap(kind error, msg string, cause error) *Error {
	return &Error{kind: kind, msg: msg, cause: cause}
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.cause)
	}
	return e.msg
}

// Is 让 errors.Is(err, ErrNotFound) 按类别命中
func (e *Error) Is(target error) bool {
	return errors.Is(e.kind, target)
}

// Unwrap 暴露底层原因，保持和 %w 包装链兼容
func (e *Error) Unwrap() error {
	return e.cause
}

// Message 取最外层领域错误的安全文案；未分类错误不应走到这里
func Message(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.msg
	}
	return "请求处理失败"
}
