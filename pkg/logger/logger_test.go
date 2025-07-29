package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	config := Config{
		Level:    "info",
		UseColor: false,
	}

	logger := New(config)

	if logger == nil {
		t.Fatal("New returned nil")
	}

	if logger.level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", logger.level)
	}

	if logger.useColor {
		t.Error("Expected useColor to be false")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"ERROR", ErrorLevel},
		{"fatal", FatalLevel},
		{"FATAL", FatalLevel},
		{"invalid", InfoLevel}, // Default
		{"", InfoLevel},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			if level != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "debug",
		Output:   &buf,
		UseColor: false,
	})

	// Test each log level
	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "[DEBUG]") {
		t.Error("Debug message not logged")
	}
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message content missing")
	}

	buf.Reset()
	logger.Info("info message")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Error("Info message not logged")
	}

	buf.Reset()
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Error("Warn message not logged")
	}

	buf.Reset()
	logger.Error("error message")
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Error("Error message not logged")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "warn",
		Output:   &buf,
		UseColor: false,
	})

	// Debug and Info should not be logged
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message should not be logged at WARN level")
	}

	logger.Info("info message")
	if buf.Len() > 0 {
		t.Error("Info message should not be logged at WARN level")
	}

	// Warn and Error should be logged
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message should be logged at WARN level")
	}

	buf.Reset()
	logger.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error message should be logged at WARN level")
	}
}

func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "debug",
		Output:   &buf,
		UseColor: false,
	})

	logger.Debugf("debug %s %d", "test", 123)
	if !strings.Contains(buf.String(), "debug test 123") {
		t.Error("Formatted debug message incorrect")
	}

	buf.Reset()
	logger.Infof("info %v", true)
	if !strings.Contains(buf.String(), "info true") {
		t.Error("Formatted info message incorrect")
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	// Create logger with field
	contextLogger := logger.WithField("user", "john")
	contextLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "user=john") {
		t.Error("Field not included in log output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Message not included in log output")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	// Create logger with multiple fields
	fields := map[string]interface{}{
		"user":   "john",
		"action": "login",
		"id":     123,
	}

	contextLogger := logger.WithFields(fields)
	contextLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "user=john") {
		t.Error("User field not included")
	}
	if !strings.Contains(output, "action=login") {
		t.Error("Action field not included")
	}
	if !strings.Contains(output, "id=123") {
		t.Error("ID field not included")
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "error",
		Output:   &buf,
		UseColor: false,
	})

	err := &testError{msg: "test error"}
	contextLogger := logger.WithError(err)
	contextLogger.Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "error=test error") {
		t.Error("Error field not included")
	}
	if !strings.Contains(output, "operation failed") {
		t.Error("Message not included")
	}
}

func TestTimeFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:      "info",
		Output:     &buf,
		UseColor:   false,
		TimeFormat: "15:04:05",
	})

	logger.Info("test")
	output := buf.String()

	// Check that output contains time in the expected format
	parts := strings.Split(output, " ")
	if len(parts) < 1 {
		t.Fatal("No output parts")
	}

	// Time should be in HH:MM:SS format
	timePart := parts[0]
	if len(timePart) != 8 || timePart[2] != ':' || timePart[5] != ':' {
		t.Errorf("Time format incorrect: %s", timePart)
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	// Debug should not be logged at info level
	logger.Debug("debug1")
	if buf.Len() > 0 {
		t.Error("Debug should not be logged at INFO level")
	}

	// Change level to debug
	logger.SetLevel(DebugLevel)

	// Now debug should be logged
	logger.Debug("debug2")
	if !strings.Contains(buf.String(), "debug2") {
		t.Error("Debug should be logged after level change")
	}
}

func TestSetOutput(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf1,
		UseColor: false,
	})

	logger.Info("message1")
	if !strings.Contains(buf1.String(), "message1") {
		t.Error("Message1 not in buf1")
	}
	if buf2.Len() > 0 {
		t.Error("buf2 should be empty")
	}

	// Change output
	logger.SetOutput(&buf2)

	logger.Info("message2")
	if !strings.Contains(buf2.String(), "message2") {
		t.Error("Message2 not in buf2")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Save original global logger
	original := globalLogger
	defer func() { globalLogger = original }()

	var buf bytes.Buffer
	testLogger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	SetGlobal(testLogger)

	// Test global functions
	Info("global info")
	if !strings.Contains(buf.String(), "global info") {
		t.Error("Global info not logged")
	}

	buf.Reset()
	Infof("global %s", "formatted")
	if !strings.Contains(buf.String(), "global formatted") {
		t.Error("Global formatted not logged")
	}
}

func TestStandardLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	stdLogger := logger.StandardLogger()
	stdLogger.Println("standard log message")

	if !strings.Contains(buf.String(), "standard log message") {
		t.Error("Standard logger message not logged")
	}
}

func TestColors(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "debug",
		Output:   &buf,
		UseColor: true,
	})

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()

	// Check for color codes
	if !strings.Contains(output, "\033[36m") { // Cyan for debug
		t.Error("Debug color not found")
	}
	if !strings.Contains(output, "\033[32m") { // Green for info
		t.Error("Info color not found")
	}
	if !strings.Contains(output, "\033[33m") { // Yellow for warn
		t.Error("Warn color not found")
	}
	if !strings.Contains(output, "\033[31m") { // Red for error
		t.Error("Error color not found")
	}
	if !strings.Contains(output, "\033[0m") { // Reset color
		t.Error("Reset color not found")
	}
}

func TestFieldsImmutability(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:    "info",
		Output:   &buf,
		UseColor: false,
	})

	// Create logger with field
	logger1 := logger.WithField("field1", "value1")
	logger2 := logger1.WithField("field2", "value2")

	// Original logger should not have fields
	buf.Reset()
	logger.Info("original")
	if strings.Contains(buf.String(), "field1") || strings.Contains(buf.String(), "field2") {
		t.Error("Original logger should not have fields")
	}

	// Logger1 should only have field1
	buf.Reset()
	logger1.Info("logger1")
	if !strings.Contains(buf.String(), "field1=value1") {
		t.Error("Logger1 should have field1")
	}
	if strings.Contains(buf.String(), "field2") {
		t.Error("Logger1 should not have field2")
	}

	// Logger2 should have both fields
	buf.Reset()
	logger2.Info("logger2")
	if !strings.Contains(buf.String(), "field1=value1") {
		t.Error("Logger2 should have field1")
	}
	if !strings.Contains(buf.String(), "field2=value2") {
		t.Error("Logger2 should have field2")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
