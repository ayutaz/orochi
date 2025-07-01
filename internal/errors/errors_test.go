package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name:     "エラーメッセージのみ",
			err:      &AppError{Code: ErrCodeNotFound, Message: "resource not found"},
			expected: "NOT_FOUND: resource not found",
		},
		{
			name:     "ラップされたエラーあり",
			err:      &AppError{Code: ErrCodeInternal, Message: "database error", Err: errors.New("connection failed")},
			expected: "INTERNAL_ERROR: database error: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	appErr := &AppError{Code: ErrCodeInternal, Message: "wrapped", Err: baseErr}

	if unwrapped := appErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Code: ErrCodeNotFound, Message: "not found"}
	err2 := &AppError{Code: ErrCodeNotFound, Message: "different message"}
	err3 := &AppError{Code: ErrCodeInternal, Message: "internal"}

	if !err1.Is(err2) {
		t.Error("errors with same code should match")
	}

	if err1.Is(err3) {
		t.Error("errors with different codes should not match")
	}

	if err1.Is(errors.New("not an AppError")) {
		t.Error("should not match non-AppError")
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() *AppError
		code     ErrorCode
		contains string
	}{
		{
			name:     "NotFound",
			fn:       func() *AppError { return NotFound("item not found") },
			code:     ErrCodeNotFound,
			contains: "item not found",
		},
		{
			name:     "NotFoundf",
			fn:       func() *AppError { return NotFoundf("item %s not found", "test") },
			code:     ErrCodeNotFound,
			contains: "item test not found",
		},
		{
			name:     "InvalidInput",
			fn:       func() *AppError { return InvalidInput("invalid data") },
			code:     ErrCodeInvalidInput,
			contains: "invalid data",
		},
		{
			name:     "InvalidInputf",
			fn:       func() *AppError { return InvalidInputf("invalid %s: %d", "port", 99999) },
			code:     ErrCodeInvalidInput,
			contains: "invalid port: 99999",
		},
		{
			name:     "Internal",
			fn:       func() *AppError { return Internal("server error") },
			code:     ErrCodeInternal,
			contains: "server error",
		},
		{
			name:     "InternalWithError",
			fn:       func() *AppError { return InternalWithError("db error", errors.New("connection lost")) },
			code:     ErrCodeInternal,
			contains: "connection lost",
		},
		{
			name:     "Conflict",
			fn:       func() *AppError { return Conflict("resource exists") },
			code:     ErrCodeConflict,
			contains: "resource exists",
		},
		{
			name:     "NetworkError",
			fn:       func() *AppError { return NetworkError("connection failed", errors.New("timeout")) },
			code:     ErrCodeNetworkError,
			contains: "timeout",
		},
		{
			name:     "ParseError",
			fn:       func() *AppError { return ParseError("invalid torrent", errors.New("bad bencode")) },
			code:     ErrCodeParseError,
			contains: "bad bencode",
		},
		{
			name:     "Timeout",
			fn:       func() *AppError { return Timeout("operation timed out") },
			code:     ErrCodeTimeout,
			contains: "operation timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()

			if err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, err.Code)
			}

			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("error message should contain %q, got %q", tt.contains, err.Error())
			}
		})
	}
}

func TestErrorCheckers(t *testing.T) {
	notFoundErr := NotFound("not found")
	invalidErr := InvalidInput("invalid")
	internalErr := Internal("internal")

	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFound with NotFound error",
			err:      notFoundErr,
			checkFn:  IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound with other error",
			err:      invalidErr,
			checkFn:  IsNotFound,
			expected: false,
		},
		{
			name:     "IsInvalidInput with InvalidInput error",
			err:      invalidErr,
			checkFn:  IsInvalidInput,
			expected: true,
		},
		{
			name:     "IsInvalidInput with other error",
			err:      notFoundErr,
			checkFn:  IsInvalidInput,
			expected: false,
		},
		{
			name:     "IsInternal with Internal error",
			err:      internalErr,
			checkFn:  IsInternal,
			expected: true,
		},
		{
			name:     "IsInternal with other error",
			err:      notFoundErr,
			checkFn:  IsInternal,
			expected: false,
		},
		{
			name:     "Checker with non-AppError",
			err:      errors.New("standard error"),
			checkFn:  IsNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFn(tt.err); got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}
