package har

import "fmt"

// ErrorCode 错误代码
type ErrorCode int

const (
	ErrInvalidFile ErrorCode = iota + 1
	ErrParseFailed
	ErrEmptyContent
	ErrDecodeFailed
	ErrWriteFailed
	ErrInvalidPath
	ErrPathTooLong
	ErrConflictResolution
)

// Error 自定义错误类型
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError 创建错误
func NewError(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
