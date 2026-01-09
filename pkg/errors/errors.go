package errors

import "fmt"

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypeDataFetch     ErrorType = "DATA_FETCH"
	ErrorTypeAnalysis      ErrorType = "ANALYSIS"
	ErrorTypeRiskControl   ErrorType = "RISK_CONTROL"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT"
	ErrorTypeCircuitBreaker ErrorType = "CIRCUIT_BREAKER"
	ErrorTypeTimeout       ErrorType = "TIMEOUT"
)

// AppError 应用错误
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新错误
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(errType ErrorType, message string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// IsType 判断错误类型
func IsType(err error, errType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errType
	}
	return false
}
