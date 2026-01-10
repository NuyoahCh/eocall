package errors

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrToolNotFound        = errors.New("tool not found")
	ErrToolExecutionFailed = errors.New("tool execution failed")
	ErrLLMRequestFailed    = errors.New("llm request failed")
	ErrRAGRetrieveFailed   = errors.New("rag retrieve failed")
	ErrInvalidInput        = errors.New("invalid input")
	ErrContextTooLong      = errors.New("context too long")
)

// AppError 应用错误
type AppError struct {
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New 创建新错误
func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装错误
func Wrap(err error, code, message string) *AppError {
	return &AppError{Code: code, Message: message, Cause: err}
}

// 错误码常量
const (
	CodeSessionNotFound = "SESSION_NOT_FOUND"
	CodeSessionExpired  = "SESSION_EXPIRED"
	CodeToolNotFound    = "TOOL_NOT_FOUND"
	CodeToolExecFailed  = "TOOL_EXEC_FAILED"
	CodeLLMFailed       = "LLM_FAILED"
	CodeRAGFailed       = "RAG_FAILED"
	CodeInvalidInput    = "INVALID_INPUT"
	CodeInternal        = "INTERNAL_ERROR"
)
