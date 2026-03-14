package errors

import (
	"context"
	"fmt"
	"log/slog"
)

// Error represents a structured application error
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Cause   error          `json:"-"`
	Context map[string]any `json:"context,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *Error) Unwrap() error {
	return e.Cause
}

// Error codes
const (
	ErrConnectionFailed = "CONNECTION_FAILED"
	ErrQueryFailed      = "QUERY_FAILED"
	ErrAuthFailed       = "AUTH_FAILED"
	ErrNotFound         = "NOT_FOUND"
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrTimeout          = "TIMEOUT"
	ErrInternal         = "INTERNAL_ERROR"
	ErrPermissionDenied = "PERMISSION_DENIED"
	ErrInvalidConfig    = "INVALID_CONFIG"
	ErrSSHFailed        = "SSH_FAILED"
	ErrSSLFailed        = "SSL_FAILED"
)

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    ErrInternal,
		Message: message,
		Cause:   err,
		Context: make(map[string]any),
	}
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    ErrInternal,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
		Context: make(map[string]any),
	}
}

// New creates a new Error with the given code and message
func New(code string, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Context: make(map[string]any),
	}
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value any) *Error {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// APIError represents a frontend-safe error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ToAPIError converts an internal error to API error
func ToAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*Error); ok {
		return &APIError{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}

	// For non-Error types, return generic error
	return &APIError{
		Code:    ErrInternal,
		Message: err.Error(),
	}
}

// ReportError logs an error with full context
func ReportError(ctx context.Context, err error, operation string) {
	if err == nil {
		return
	}

	if appErr, ok := err.(*Error); ok {
		slog.ErrorContext(ctx, operation+" failed",
			"code", appErr.Code,
			"message", appErr.Message,
			"cause", appErr.Cause,
			"context", appErr.Context,
		)
	} else {
		slog.ErrorContext(ctx, operation+" failed",
			"error", err,
		)
	}
}
