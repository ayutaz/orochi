package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  Level
		msgLevel  Level
		msg       string
		shouldLog bool
	}{
		{"Debug at Debug level", DebugLevel, DebugLevel, "debug message", true},
		{"Debug at Info level", InfoLevel, DebugLevel, "debug message", false},
		{"Info at Info level", InfoLevel, InfoLevel, "info message", true},
		{"Warn at Info level", InfoLevel, WarnLevel, "warn message", true},
		{"Error at Warn level", WarnLevel, ErrorLevel, "error message", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &Config{
				Level:      tt.logLevel,
				Output:     &buf,
				TimeFormat: "2006-01-02",
			}

			log := New(config)

			switch tt.msgLevel {
			case DebugLevel:
				log.Debug(tt.msg)
			case InfoLevel:
				log.Info(tt.msg)
			case WarnLevel:
				log.Warn(tt.msg)
			case ErrorLevel:
				log.Error(tt.msg)
			}

			output := buf.String()
			if tt.shouldLog && output == "" {
				t.Error("expected log output but got none")
			}
			if !tt.shouldLog && output != "" {
				t.Errorf("expected no log output but got: %s", output)
			}
			if tt.shouldLog && !strings.Contains(output, tt.msg) {
				t.Errorf("log output missing message: %s", output)
			}
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:      InfoLevel,
		Output:     &buf,
		TimeFormat: "2006-01-02",
	}

	log := New(config)
	logWithFields := log.WithFields(
		String("component", "test"),
		Int("count", 42),
	)

	logWithFields.Info("test message", String("extra", "field"))

	output := buf.String()
	if !strings.Contains(output, "component") {
		t.Error("output missing component field")
	}
	if !strings.Contains(output, "42") {
		t.Error("output missing count field")
	}
	if !strings.Contains(output, "extra") {
		t.Error("output missing extra field")
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:      InfoLevel,
		Output:     &buf,
		TimeFormat: "2006-01-02",
	}

	log := New(config)
	ctx := context.WithValue(context.Background(), requestIDKey, "test-123")
	logWithCtx := log.WithContext(ctx)

	logWithCtx.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Error("output missing request_id from context")
	}
	if !strings.Contains(output, "test-123") {
		t.Error("output missing request_id value")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{Level(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		key   string
		value interface{}
	}{
		{"String field", String("key", "value"), "key", "value"},
		{"Int field", Int("count", 42), "count", 42},
		{"Int64 field", Int64("big", int64(9999)), "big", int64(9999)},
		{"Bool field", Bool("flag", true), "flag", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Key != tt.key {
				t.Errorf("expected key %s, got %s", tt.key, tt.field.Key)
			}
			if tt.field.Value != tt.value {
				t.Errorf("expected value %v, got %v", tt.value, tt.field.Value)
			}
		})
	}
}

func TestErr(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		field := Err(nil)
		if field.Key != "error" {
			t.Errorf("expected key 'error', got %s", field.Key)
		}
		if field.Value != nil {
			t.Errorf("expected nil value, got %v", field.Value)
		}
	})

	t.Run("non-nil error", func(t *testing.T) {
		_, err := strings.NewReader("").Read(nil)
		field := Err(err)
		if field.Key != "error" {
			t.Errorf("expected key 'error', got %s", field.Key)
		}
		if field.Value == nil {
			t.Error("expected non-nil value")
		}
	})
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Level:      InfoLevel,
		Output:     &buf,
		TimeFormat: "2006-01-02",
	}

	SetGlobal(New(config))

	Info("global test")

	output := buf.String()
	if !strings.Contains(output, "global test") {
		t.Error("global logger not working")
	}
}
