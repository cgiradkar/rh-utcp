package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New(ErrorTypeValidation, "test error")

	if err == nil {
		t.Fatal("New returned nil")
	}

	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", err.Message)
	}

	if len(err.Stack) == 0 {
		t.Error("Stack trace not captured")
	}

	if err.Context == nil {
		t.Error("Context not initialized")
	}
}

func TestNewfError(t *testing.T) {
	err := Newf(ErrorTypeNotFound, "resource %s not found", "user")

	if err.Message != "resource user not found" {
		t.Errorf("Expected formatted message, got %s", err.Message)
	}
}

func TestErrorInterface(t *testing.T) {
	err := New(ErrorTypeInternal, "internal error")

	// Test Error() method
	errStr := err.Error()
	expected := "internal: internal error"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}

	// Test with cause
	cause := errors.New("root cause")
	wrapped := Wrap(cause, ErrorTypeInternal, "wrapped error")

	errStr = wrapped.Error()
	expected = "internal: wrapped error: root cause"
	if errStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errStr)
	}
}

func TestWrap(t *testing.T) {
	// Test wrapping nil
	wrapped := Wrap(nil, ErrorTypeInternal, "should be nil")
	if wrapped != nil {
		t.Error("Wrapping nil should return nil")
	}

	// Test wrapping standard error
	stdErr := errors.New("standard error")
	wrapped = Wrap(stdErr, ErrorTypeProvider, "provider failed")

	if wrapped.Type != ErrorTypeProvider {
		t.Errorf("Expected type %s, got %s", ErrorTypeProvider, wrapped.Type)
	}

	if wrapped.Message != "provider failed" {
		t.Errorf("Expected message 'provider failed', got %s", wrapped.Message)
	}

	if wrapped.Cause != stdErr {
		t.Error("Cause not preserved")
	}

	// Test wrapping our error type
	ourErr := New(ErrorTypeValidation, "validation error")
	ourErr.Provider = "test-provider"
	ourErr.Operation = "test-op"

	wrapped = Wrap(ourErr, ErrorTypeInternal, "internal wrapper")

	// Should preserve original error properties
	if wrapped.Provider != "test-provider" {
		t.Error("Provider not preserved when wrapping our error type")
	}

	if wrapped.Operation != "test-op" {
		t.Error("Operation not preserved when wrapping our error type")
	}
}

func TestWrapf(t *testing.T) {
	err := errors.New("base error")
	wrapped := Wrapf(err, ErrorTypeNetwork, "network error: %s", "timeout")

	if wrapped.Message != "network error: timeout" {
		t.Errorf("Expected formatted message, got %s", wrapped.Message)
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	wrapped := Wrap(cause, ErrorTypeInternal, "wrapped")

	unwrapped := wrapped.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap did not return original cause")
	}
}

func TestWithContext(t *testing.T) {
	err := New(ErrorTypeValidation, "test")

	err.WithContext("field", "username")
	err.WithContext("value", "invalid")

	if err.Context["field"] != "username" {
		t.Error("Field context not set")
	}

	if err.Context["value"] != "invalid" {
		t.Error("Value context not set")
	}
}

func TestWithProvider(t *testing.T) {
	// Test with standard error
	stdErr := errors.New("standard error")
	err := WithProvider(stdErr, "jira")

	if err.Provider != "jira" {
		t.Errorf("Expected provider 'jira', got %s", err.Provider)
	}

	// Test with our error type
	ourErr := New(ErrorTypeProvider, "provider error")
	err = WithProvider(ourErr, "gitlab")

	if err.Provider != "gitlab" {
		t.Errorf("Expected provider 'gitlab', got %s", err.Provider)
	}
}

func TestWithOperation(t *testing.T) {
	err := New(ErrorTypeNetwork, "network error")
	err = WithOperation(err, "fetch_data")

	if err.Operation != "fetch_data" {
		t.Errorf("Expected operation 'fetch_data', got %s", err.Operation)
	}
}

func TestWithStatusCode(t *testing.T) {
	err := New(ErrorTypeNotFound, "not found")
	err = WithStatusCode(err, 404)

	if err.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", err.StatusCode)
	}
}

func TestIs(t *testing.T) {
	// Test with nil
	if Is(nil, ErrorTypeValidation) {
		t.Error("Is should return false for nil error")
	}

	// Test with our error type
	err := New(ErrorTypeValidation, "validation error")
	if !Is(err, ErrorTypeValidation) {
		t.Error("Is should return true for matching type")
	}

	if Is(err, ErrorTypeNotFound) {
		t.Error("Is should return false for non-matching type")
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if Is(stdErr, ErrorTypeValidation) {
		t.Error("Is should return false for standard error")
	}
}

func TestGetType(t *testing.T) {
	// Test with nil
	if GetType(nil) != "" {
		t.Error("GetType should return empty string for nil")
	}

	// Test with our error
	err := New(ErrorTypeTimeout, "timeout")
	if GetType(err) != ErrorTypeTimeout {
		t.Errorf("Expected type %s, got %s", ErrorTypeTimeout, GetType(err))
	}

	// Test with standard error
	stdErr := errors.New("standard")
	if GetType(stdErr) != ErrorTypeInternal {
		t.Errorf("Expected type %s for standard error, got %s", ErrorTypeInternal, GetType(stdErr))
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		errorType    ErrorType
		expectedCode int
	}{
		{ErrorTypeValidation, 400},
		{ErrorTypeNotFound, 404},
		{ErrorTypeUnauthorized, 401},
		{ErrorTypeForbidden, 403},
		{ErrorTypeTimeout, 408},
		{ErrorTypeConfiguration, 500},
		{ErrorTypeProvider, 502},
		{ErrorTypeNetwork, 503},
		{ErrorTypeInternal, 500},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			err := New(tt.errorType, "test")
			code := GetStatusCode(err)
			if code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, code)
			}
		})
	}

	// Test with custom status code
	err := New(ErrorTypeValidation, "test")
	err = WithStatusCode(err, 422)
	if GetStatusCode(err) != 422 {
		t.Error("Custom status code not returned")
	}

	// Test with nil
	if GetStatusCode(nil) != 200 {
		t.Error("GetStatusCode should return 200 for nil")
	}

	// Test with standard error
	if GetStatusCode(errors.New("test")) != 500 {
		t.Error("GetStatusCode should return 500 for standard error")
	}
}

func TestGetStack(t *testing.T) {
	// Test with nil
	if GetStack(nil) != nil {
		t.Error("GetStack should return nil for nil error")
	}

	// Test with our error
	err := New(ErrorTypeInternal, "test")
	stack := GetStack(err)
	if len(stack) == 0 {
		t.Error("GetStack should return non-empty stack")
	}

	// Test with standard error
	if GetStack(errors.New("test")) != nil {
		t.Error("GetStack should return nil for standard error")
	}
}

func TestFormatStack(t *testing.T) {
	// Test with empty stack
	formatted := FormatStack(nil)
	if formatted != "" {
		t.Error("FormatStack should return empty string for nil stack")
	}

	// Test with stack
	stack := []StackFrame{
		{
			Function: "main.testFunc",
			File:     "/path/to/file.go",
			Line:     42,
		},
	}

	formatted = FormatStack(stack)
	if formatted == "" {
		t.Error("FormatStack should return non-empty string")
	}

	if !contains(formatted, "main.testFunc") {
		t.Error("Formatted stack should contain function name")
	}

	if !contains(formatted, "/path/to/file.go:42") {
		t.Error("Formatted stack should contain file and line")
	}
}

func TestCommonConstructors(t *testing.T) {
	// ValidationError
	err := ValidationError("invalid input")
	if err.Type != ErrorTypeValidation {
		t.Error("ValidationError should have validation type")
	}

	// ValidationErrorf
	err = ValidationErrorf("field %s is required", "email")
	if err.Message != "field email is required" {
		t.Error("ValidationErrorf formatting incorrect")
	}

	// NotFoundError
	err = NotFoundError("user")
	if err.Message != "user not found" {
		t.Error("NotFoundError message incorrect")
	}

	// UnauthorizedError
	err = UnauthorizedError("invalid token")
	if err.Type != ErrorTypeUnauthorized {
		t.Error("UnauthorizedError type incorrect")
	}

	// ForbiddenError
	err = ForbiddenError("access denied")
	if err.Type != ErrorTypeForbidden {
		t.Error("ForbiddenError type incorrect")
	}

	// InternalError
	err = InternalError("server error")
	if err.Type != ErrorTypeInternal {
		t.Error("InternalError type incorrect")
	}

	// InternalErrorf
	err = InternalErrorf("failed to %s", "connect")
	if err.Message != "failed to connect" {
		t.Error("InternalErrorf formatting incorrect")
	}

	// ConfigurationError
	err = ConfigurationError("missing config")
	if err.Type != ErrorTypeConfiguration {
		t.Error("ConfigurationError type incorrect")
	}

	// ConfigurationErrorf
	err = ConfigurationErrorf("invalid %s value", "port")
	if err.Message != "invalid port value" {
		t.Error("ConfigurationErrorf formatting incorrect")
	}

	// ProviderError
	err = ProviderError("jira", "connection failed")
	if err.Type != ErrorTypeProvider {
		t.Error("ProviderError type incorrect")
	}
	if err.Context["provider"] != "jira" {
		t.Error("ProviderError context incorrect")
	}

	// ProviderErrorf
	err = ProviderErrorf("gitlab", "API returned %d", 500)
	if err.Message != "API returned 500" {
		t.Error("ProviderErrorf formatting incorrect")
	}

	// NetworkError
	err = NetworkError("connection timeout")
	if err.Type != ErrorTypeNetwork {
		t.Error("NetworkError type incorrect")
	}

	// TimeoutError
	err = TimeoutError("fetch_data")
	if err.Message != "operation timed out: fetch_data" {
		t.Error("TimeoutError message incorrect")
	}
}

func TestStackCapture(t *testing.T) {
	err := New(ErrorTypeInternal, "test")

	if len(err.Stack) == 0 {
		t.Fatal("Stack not captured")
	}

	// First frame should be from this test
	firstFrame := err.Stack[0]
	if !contains(firstFrame.Function, "TestStackCapture") {
		t.Error("Stack should contain current test function")
	}

	// Stack should be limited to 10 frames
	if len(err.Stack) > 10 {
		t.Error("Stack should be limited to 10 frames")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
