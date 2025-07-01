package errors

import "fmt"

// ErrorCode represents the type of error.
type ErrorCode string

const (
	// ErrCodeNotFound indicates that a resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"

	// ErrCodeInvalidInput indicates that the input provided was invalid.
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"

	// ErrCodeUnauthorized indicates unauthorized access.
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"

	// ErrCodeConflict indicates a resource conflict.
	ErrCodeConflict ErrorCode = "CONFLICT"

	// ErrCodeTimeout indicates an operation timeout.
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeNetworkError indicates a network-related error.
	ErrCodeNetworkError ErrorCode = "NETWORK_ERROR"

	// ErrCodeParseError indicates a parsing error.
	ErrCodeParseError ErrorCode = "PARSE_ERROR"

	// ErrCodePermissionDenied indicates permission denied.
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
)

// AppError represents a structured application error.
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error returns the error message.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is implements error comparison.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NotFound creates a new NOT_FOUND error.
func NotFound(message string) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: message}
}

// NotFoundf creates a new NOT_FOUND error with formatting.
func NotFoundf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeNotFound, Message: fmt.Sprintf(format, args...)}
}

// InvalidInput creates a new INVALID_INPUT error.
func InvalidInput(message string) *AppError {
	return &AppError{Code: ErrCodeInvalidInput, Message: message}
}

// InvalidInputf creates a new INVALID_INPUT error with formatting.
func InvalidInputf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeInvalidInput, Message: fmt.Sprintf(format, args...)}
}

// Internal creates a new INTERNAL_ERROR error.
func Internal(message string) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: message}
}

// InternalWithError creates a new INTERNAL_ERROR error with wrapped error.
func InternalWithError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: message, Err: err}
}

// Conflict creates a new CONFLICT error.
func Conflict(message string) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: message}
}

// NetworkError creates a new NETWORK_ERROR error.
func NetworkError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeNetworkError, Message: message, Err: err}
}

// ParseError creates a new PARSE_ERROR error.
func ParseError(message string, err error) *AppError {
	return &AppError{Code: ErrCodeParseError, Message: message, Err: err}
}

// Timeout creates a new TIMEOUT error.
func Timeout(message string) *AppError {
	return &AppError{Code: ErrCodeTimeout, Message: message}
}

// IsNotFound checks if an error is a NOT_FOUND error.
func IsNotFound(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Code == ErrCodeNotFound
}

// IsInvalidInput checks if an error is an INVALID_INPUT error.
func IsInvalidInput(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Code == ErrCodeInvalidInput
}

// IsInternal checks if an error is an INTERNAL_ERROR error.
func IsInternal(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Code == ErrCodeInternal
}

// IsParseError checks if an error is a PARSE_ERROR error.
func IsParseError(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Code == ErrCodeParseError
}

// AlreadyExistsf creates a new CONFLICT error with formatting.
func AlreadyExistsf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeConflict, Message: fmt.Sprintf(format, args...)}
}

// InternalErrorf creates a new INTERNAL_ERROR error with formatting.
func InternalErrorf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeInternal, Message: fmt.Sprintf(format, args...)}
}

// AuthenticationFailedf creates a new UNAUTHORIZED error with formatting.
func AuthenticationFailedf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeUnauthorized, Message: fmt.Sprintf(format, args...)}
}

// IsAuthenticationFailed checks if an error is an UNAUTHORIZED error.
func IsAuthenticationFailed(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Code == ErrCodeUnauthorized
}

// ValidationErrorf creates a new INVALID_INPUT error with formatting.
func ValidationErrorf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodeInvalidInput, Message: fmt.Sprintf(format, args...)}
}

// PermissionDeniedf creates a new PERMISSION_DENIED error with formatting.
func PermissionDeniedf(format string, args ...interface{}) *AppError {
	return &AppError{Code: ErrCodePermissionDenied, Message: fmt.Sprintf(format, args...)}
}
