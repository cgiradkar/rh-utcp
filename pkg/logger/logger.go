package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DebugLevel logs everything
	DebugLevel LogLevel = iota
	// InfoLevel logs info, warnings, and errors
	InfoLevel
	// WarnLevel logs warnings and errors
	WarnLevel
	// ErrorLevel logs only errors
	ErrorLevel
	// FatalLevel logs fatal errors and exits
	FatalLevel
)

var (
	levelNames = map[LogLevel]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
		FatalLevel: "FATAL",
	}

	levelColors = map[LogLevel]string{
		DebugLevel: "\033[36m", // Cyan
		InfoLevel:  "\033[32m", // Green
		WarnLevel:  "\033[33m", // Yellow
		ErrorLevel: "\033[31m", // Red
		FatalLevel: "\033[35m", // Magenta
	}

	resetColor = "\033[0m"
)

// Logger is the main logger interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// StructuredLogger implements the Logger interface
type StructuredLogger struct {
	mu         sync.RWMutex
	level      LogLevel
	output     io.Writer
	fields     map[string]interface{}
	useColor   bool
	showCaller bool
	timeFormat string
}

// Config holds logger configuration
type Config struct {
	Level      string
	Output     io.Writer
	UseColor   bool
	ShowCaller bool
	TimeFormat string
}

// New creates a new logger instance
func New(config Config) *StructuredLogger {
	level := parseLevel(config.Level)

	output := config.Output
	if output == nil {
		output = os.Stdout
	}

	timeFormat := config.TimeFormat
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}

	return &StructuredLogger{
		level:      level,
		output:     output,
		fields:     make(map[string]interface{}),
		useColor:   config.UseColor,
		showCaller: config.ShowCaller,
		timeFormat: timeFormat,
	}
}

// Default creates a logger with default settings
func Default() *StructuredLogger {
	return New(Config{
		Level:    "info",
		UseColor: true,
	})
}

// parseLevel converts a string level to LogLevel
func parseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// SetLevel sets the logging level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *StructuredLogger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// log is the internal logging method
func (l *StructuredLogger) log(level LogLevel, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// Build the log entry
	entry := l.formatEntry(level, fmt.Sprint(args...))

	// Write to output
	fmt.Fprint(l.output, entry)

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// logf is the internal formatted logging method
func (l *StructuredLogger) logf(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// Build the log entry
	entry := l.formatEntry(level, fmt.Sprintf(format, args...))

	// Write to output
	fmt.Fprint(l.output, entry)

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// formatEntry formats a log entry
func (l *StructuredLogger) formatEntry(level LogLevel, message string) string {
	var parts []string

	// Timestamp
	parts = append(parts, time.Now().Format(l.timeFormat))

	// Level
	levelStr := levelNames[level]
	if l.useColor {
		levelStr = levelColors[level] + levelStr + resetColor
	}
	parts = append(parts, fmt.Sprintf("[%s]", levelStr))

	// Caller information
	if l.showCaller {
		_, file, line, ok := runtime.Caller(4)
		if ok {
			// Get just the filename without the full path
			parts = append(parts, fmt.Sprintf("%s:%d", filepath.Base(file), line))
		}
	}

	// Fields
	if len(l.fields) > 0 {
		var fieldParts []string
		for k, v := range l.fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	// Message
	parts = append(parts, message)

	return strings.Join(parts, " ") + "\n"
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(args ...interface{}) {
	l.log(DebugLevel, args...)
}

// Debugf logs a formatted debug message
func (l *StructuredLogger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, format, args...)
}

// Info logs an info message
func (l *StructuredLogger) Info(args ...interface{}) {
	l.log(InfoLevel, args...)
}

// Infof logs a formatted info message
func (l *StructuredLogger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(args ...interface{}) {
	l.log(WarnLevel, args...)
}

// Warnf logs a formatted warning message
func (l *StructuredLogger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, format, args...)
}

// Error logs an error message
func (l *StructuredLogger) Error(args ...interface{}) {
	l.log(ErrorLevel, args...)
}

// Errorf logs a formatted error message
func (l *StructuredLogger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, format, args...)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(args ...interface{}) {
	l.log(FatalLevel, args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *StructuredLogger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
}

// WithField creates a new logger with an additional field
func (l *StructuredLogger) WithField(key string, value interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Copy fields
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &StructuredLogger{
		level:      l.level,
		output:     l.output,
		fields:     newFields,
		useColor:   l.useColor,
		showCaller: l.showCaller,
		timeFormat: l.timeFormat,
	}
}

// WithFields creates a new logger with additional fields
func (l *StructuredLogger) WithFields(fields map[string]interface{}) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Copy fields
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &StructuredLogger{
		level:      l.level,
		output:     l.output,
		fields:     newFields,
		useColor:   l.useColor,
		showCaller: l.showCaller,
		timeFormat: l.timeFormat,
	}
}

// WithError creates a new logger with an error field
func (l *StructuredLogger) WithError(err error) Logger {
	return l.WithField("error", err.Error())
}

// Global logger instance
var globalLogger = Default()

// SetGlobal sets the global logger
func SetGlobal(logger *StructuredLogger) {
	globalLogger = logger
}

// GetGlobal returns the global logger
func GetGlobal() *StructuredLogger {
	return globalLogger
}

// Package-level convenience functions

// Debug logs a debug message using the global logger
func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs an info message using the global logger
func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs an error message using the global logger
func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(args ...interface{}) {
	globalLogger.Fatal(args...)
}

// Fatalf logs a formatted fatal message using the global logger and exits
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// StandardLogger returns a standard library logger that writes to this logger
func (l *StructuredLogger) StandardLogger() *log.Logger {
	return log.New(&logWriter{logger: l, level: InfoLevel}, "", 0)
}

// logWriter adapts our logger to io.Writer interface
type logWriter struct {
	logger *StructuredLogger
	level  LogLevel
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.log(w.level, strings.TrimSpace(string(p)))
	return len(p), nil
}
