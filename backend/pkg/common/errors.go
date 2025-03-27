package common

import (
	"fmt"
)

// AppError represents an application-specific error
type AppError struct {
	Message string
	Code    string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError creates a new AppError with just a message
func NewError(message string) *AppError {
	return &AppError{
		Message: message,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, message string) *AppError {
	return &AppError{
		Message: message,
		Err:     err,
	}
}

// WithCode adds an error code to the AppError
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}