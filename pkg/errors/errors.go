package errors

import (
	"fmt"
	"runtime"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeUnauthorized indicates an authorization error
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden indicates a forbidden access error
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeConfiguration indicates a configuration error
	ErrorTypeConfiguration ErrorType = "configuration"
	// ErrorTypeProvider indicates a provider-specific error
	ErrorTypeProvider ErrorType = "provider"
	// ErrorTypeNetwork indicates a network-related error
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeTimeout indicates a timeout error
	ErrorTypeTimeout ErrorType = "timeout"
)

// Error represents a structured error with additional context
type Error struct {
	Type       ErrorType
	Message    string
	Provider   string
	Operation  string
	StatusCode int
	Cause      error
	Stack      []StackFrame
	Context    map[string]interface{}
}

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new error
func New(errorType ErrorType, message string) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Stack:   captureStack(2),
		Context: make(map[string]interface{}),
	}
}

// Newf creates a new formatted error
func Newf(errorType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
		Stack:   captureStack(2),
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *Error {
	if err == nil {
		return nil
	}

	// If it's already our error type, preserve the stack
	if e, ok := err.(*Error); ok {
		return &Error{
			Type:       errorType,
			Message:    message,
			Provider:   e.Provider,
			Operation:  e.Operation,
			StatusCode: e.StatusCode,
			Cause:      err,
			Stack:      e.Stack,
			Context:    e.Context,
		}
	}

	return &Error{
		Type:    errorType,
		Message: message,
		Cause:   err,
		Stack:   captureStack(2),
		Context: make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}

	return Wrap(err, errorType, fmt.Sprintf(format, args...))
}

// WithProvider adds provider information to the error
func WithProvider(err error, provider string) *Error {
	e := ensureError(err)
	e.Provider = provider
	return e
}

// WithOperation adds operation information to the error
func WithOperation(err error, operation string) *Error {
	e := ensureError(err)
	e.Operation = operation
	return e
}

// WithStatusCode adds HTTP status code to the error
func WithStatusCode(err error, statusCode int) *Error {
	e := ensureError(err)
	e.StatusCode = statusCode
	return e
}

// ensureError ensures we have an *Error type
func ensureError(err error) *Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(*Error); ok {
		return e
	}

	return &Error{
		Type:    ErrorTypeInternal,
		Message: err.Error(),
		Cause:   err,
		Stack:   captureStack(3),
		Context: make(map[string]interface{}),
	}
}

// captureStack captures the current stack trace
func captureStack(skip int) []StackFrame {
	var frames []StackFrame

	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})

		// Limit stack depth
		if len(frames) >= 10 {
			break
		}
	}

	return frames
}

// Is checks if an error is of a specific type
func Is(err error, errorType ErrorType) bool {
	if err == nil {
		return false
	}

	e, ok := err.(*Error)
	if !ok {
		return false
	}

	return e.Type == errorType
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if err == nil {
		return ""
	}

	e, ok := err.(*Error)
	if !ok {
		return ErrorTypeInternal
	}

	return e.Type
}

// GetStatusCode returns the HTTP status code for the error
func GetStatusCode(err error) int {
	if err == nil {
		return 200
	}

	e, ok := err.(*Error)
	if !ok {
		return 500
	}

	if e.StatusCode != 0 {
		return e.StatusCode
	}

	// Default status codes based on error type
	switch e.Type {
	case ErrorTypeValidation:
		return 400
	case ErrorTypeNotFound:
		return 404
	case ErrorTypeUnauthorized:
		return 401
	case ErrorTypeForbidden:
		return 403
	case ErrorTypeTimeout:
		return 408
	case ErrorTypeConfiguration:
		return 500
	case ErrorTypeProvider:
		return 502
	case ErrorTypeNetwork:
		return 503
	default:
		return 500
	}
}

// GetStack returns the stack trace from an error
func GetStack(err error) []StackFrame {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return nil
	}

	return e.Stack
}

// FormatStack formats a stack trace as a string
func FormatStack(frames []StackFrame) string {
	if len(frames) == 0 {
		return ""
	}

	var result string
	for _, frame := range frames {
		result += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
	}

	return result
}

// Common error constructors

// ValidationError creates a validation error
func ValidationError(message string) *Error {
	return New(ErrorTypeValidation, message)
}

// ValidationErrorf creates a formatted validation error
func ValidationErrorf(format string, args ...interface{}) *Error {
	return Newf(ErrorTypeValidation, format, args...)
}

// NotFoundError creates a not found error
func NotFoundError(resource string) *Error {
	return Newf(ErrorTypeNotFound, "%s not found", resource)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *Error {
	return New(ErrorTypeUnauthorized, message)
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *Error {
	return New(ErrorTypeForbidden, message)
}

// InternalError creates an internal error
func InternalError(message string) *Error {
	return New(ErrorTypeInternal, message)
}

// InternalErrorf creates a formatted internal error
func InternalErrorf(format string, args ...interface{}) *Error {
	return Newf(ErrorTypeInternal, format, args...)
}

// ConfigurationError creates a configuration error
func ConfigurationError(message string) *Error {
	return New(ErrorTypeConfiguration, message)
}

// ConfigurationErrorf creates a formatted configuration error
func ConfigurationErrorf(format string, args ...interface{}) *Error {
	return Newf(ErrorTypeConfiguration, format, args...)
}

// ProviderError creates a provider error
func ProviderError(provider, message string) *Error {
	e := New(ErrorTypeProvider, message)
	e.WithContext("provider", provider)
	return e
}

// ProviderErrorf creates a formatted provider error
func ProviderErrorf(provider, format string, args ...interface{}) *Error {
	e := Newf(ErrorTypeProvider, format, args...)
	e.WithContext("provider", provider)
	return e
}

// NetworkError creates a network error
func NetworkError(message string) *Error {
	return New(ErrorTypeNetwork, message)
}

// TimeoutError creates a timeout error
func TimeoutError(operation string) *Error {
	return Newf(ErrorTypeTimeout, "operation timed out: %s", operation)
}
