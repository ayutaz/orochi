package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	// DebugLevel is the most verbose level.
	DebugLevel Level = iota
	// InfoLevel is for informational messages.
	InfoLevel
	// WarnLevel is for warning messages.
	WarnLevel
	// ErrorLevel is for error messages.
	ErrorLevel
	// FatalLevel is for fatal errors that cause the program to exit.
	FatalLevel
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Value interface{}
	Key   string
}

// Logger defines the interface for logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Config represents logger configuration.
type Config struct {
	Output     io.Writer
	TimeFormat string
	Level      Level
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
	}
}

// logger is the default implementation of Logger.
type logger struct {
	context context.Context
	config  *Config
	fields  []Field
}

// New creates a new logger with the given configuration.
func New(config *Config) Logger {
	if config == nil {
		config = DefaultConfig()
	}
	return &logger{
		config: config,
		fields: make([]Field, 0),
	}
}

// NewWithLevel creates a new logger with the specified level.
func NewWithLevel(level Level) Logger {
	config := DefaultConfig()
	config.Level = level
	return New(config)
}

func (l *logger) log(level Level, msg string, fields ...Field) {
	if level < l.config.Level {
		return
	}

	// Combine logger fields with message fields
	allFields := make([]Field, 0, len(l.fields)+len(fields)+3)

	// Add standard fields
	allFields = append(allFields,
		Field{Key: "time", Value: time.Now().Format(l.config.TimeFormat)},
		Field{Key: "level", Value: level.String()},
		Field{Key: "msg", Value: msg},
	)

	// Add logger fields
	allFields = append(allFields, l.fields...)

	// Add message fields
	allFields = append(allFields, fields...)

	// Add context fields if available
	if l.context != nil {
		if reqID := l.context.Value(requestIDKey); reqID != nil {
			allFields = append(allFields, Field{Key: "request_id", Value: reqID})
		}
	}

	// Format and write the log entry
	entry := l.formatEntry(allFields)
	fmt.Fprintln(l.config.Output, entry)

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

func (l *logger) formatEntry(fields []Field) string {
	// Simple JSON-like format
	result := "{"
	for i, field := range fields {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`%q: "%v"`, field.Key, field.Value)
	}
	result += "}"
	return result
}

func (l *logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, fields...)
}

func (l *logger) WithFields(fields ...Field) Logger {
	newLogger := &logger{
		config:  l.config,
		fields:  make([]Field, len(l.fields)+len(fields)),
		context: l.context,
	}
	copy(newLogger.fields, l.fields)
	copy(newLogger.fields[len(l.fields):], fields)
	return newLogger
}

func (l *logger) WithContext(ctx context.Context) Logger {
	return &logger{
		config:  l.config,
		fields:  l.fields,
		context: ctx,
	}
}

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for request ID.
	requestIDKey contextKey = "request_id"
)

// Global logger instance.
var globalLogger = New(DefaultConfig())

// SetGlobal sets the global logger.
func SetGlobal(l Logger) {
	globalLogger = l
}

// Global returns the global logger.
func Global() Logger {
	return globalLogger
}

// Convenience functions using the global logger.
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	globalLogger.Fatal(msg, fields...)
}

// Helper functions for common fields.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
