package errors

import (
	"fmt"
	"runtime"
)

// ErrorCode 定义错误码类型
// ErrorCode defines the type for error codes.
type ErrorCode string

// 通用错误码定义
// Common error code definitions.
const (
	UnknownError         ErrorCode = "UnknownError"         // 未知错误 Unknown error
	NotFoundError        ErrorCode = "NotFoundError"        // 资源未找到 Resource not found
	InvalidArgument      ErrorCode = "InvalidArgument"      // 无效参数 Invalid argument
	PermissionDenied     ErrorCode = "PermissionDenied"     // 权限不足 Permission denied
	RateLimitExceeded    ErrorCode = "RateLimitExceeded"    // 超出速率限制 Rate limit exceeded
	InternalError        ErrorCode = "InternalError"        // 内部错误 Internal server error
	DatabaseError        ErrorCode = "DatabaseError"        // 数据库错误 Database error
	NetworkError         ErrorCode = "NetworkError"         // 网络错误 Network error
	ConfigError          ErrorCode = "ConfigError"          // 配置错误 Configuration error
	SerializationError   ErrorCode = "SerializationError"   // 序列化错误 Serialization error
	DeserializationError ErrorCode = "DeserializationError" // 反序列化错误 Deserialization error
	TimeoutError         ErrorCode = "TimeoutError"         // 操作超时 Operation timed out
	AlreadyExistsError   ErrorCode = "AlreadyExistsError"   // 资源已存在 Resource already exists
)

// AppError 是应用程序的自定义错误结构
// AppError is the custom error structure for the application.
type AppError struct {
	Code       ErrorCode // 错误码 Error code
	Message    string    // 错误信息 Error message
	StackTrace string    // 堆栈信息 Stack trace
	Err        error     // 原始错误 Original error (optional)
}

// Error 实现 error 接口
// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Code: %s, Message: %s, OriginalError: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("Code: %s, Message: %s", e.Code, e.Message)
}

// Unwrap 提供对原始错误的支持，用于 errors.Is 和 errors.As
// Unwrap provides support for the original error, for use with errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建一个新的 AppError
// New creates a new AppError.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StackTrace: getStackTrace(),
	}
}

// Newf 创建一个带格式化消息的 AppError
// Newf creates a new AppError with a formatted message.
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: getStackTrace(),
	}
}

// Wrap 包装一个现有错误为 AppError
// Wrap wraps an existing error into an AppError.
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return New(code, message)
	}
	// If err is already an AppError, don't double wrap its specific message/code,
	// but preserve its stack trace and original error if different.
	if appErr, ok := err.(*AppError); ok {
		// If the new message or code is more specific, use it.
		// This behavior might need adjustment based on desired wrapping strategy.
		// For now, we just re-wrap with new context.
		return &AppError{
			Code:       code, // Potentially overwriting a more specific code from appErr
			Message:    message,
			StackTrace: appErr.StackTrace, // Preserve original stack trace
			Err:        appErr,            // Chain the original AppError
		}
	}
	return &AppError{
		Code:       code,
		Message:    message,
		StackTrace: getStackTrace(),
		Err:        err,
	}
}

// Wrapf 包装一个现有错误为 AppError，并使用格式化消息
// Wrapf wraps an existing error into an AppError with a formatted message.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	if err == nil {
		return Newf(code, format, args...)
	}
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:       code,
			Message:    fmt.Sprintf(format, args...),
			StackTrace: appErr.StackTrace,
			Err:        appErr,
		}
	}
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: getStackTrace(),
		Err:        err,
	}
}

// GetCode 从错误中获取错误码，如果错误不是 AppError 则返回 UnknownError
// GetCode retrieves the error code from an error. Returns UnknownError if not an AppError.
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Code
	}
	return UnknownError
}

// GetMessage 从错误中获取消息
// GetMessage retrieves the message from an error.
func GetMessage(err error) string {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}

// Is 检查错误是否是特定的 AppError 类型（通过错误码比较）
// Is checks if an error is of a specific AppError type (by comparing error codes).
func Is(err error, targetCode ErrorCode) bool {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.Code == targetCode
	}
	return false
}

// As 类似于标准库的 errors.As，但专门用于 *AppError
// As is similar to the standard library's errors.As, but specialized for *AppError.
func As(err error, target **AppError) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*AppError)
	if ok {
		*target = e
		return true
	}
	// Check wrapped errors
	cause := e
	for cause.Err != nil {
		if e, ok := cause.Err.(*AppError); ok {
			*target = e
			return true
		}
		if unwrapper, ok := cause.Err.(interface{ Unwrap() error }); ok {
			cause = Wrap(unwrapper.Unwrap(), cause.Code, cause.Message).(*AppError) // temporary wrap to continue loop with AppError structure
		} else {
			break
		}
	}
	return false
}

// getStackTrace 获取当前的堆栈信息
// getStackTrace retrieves the current stack trace.
func getStackTrace() string {
	buf := make([]byte, 1024) // 可以根据需要调整大小 Can be adjusted as needed
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
