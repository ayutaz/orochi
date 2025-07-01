package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"github.com/ayutaz/orochi/internal/logger"
)

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&logger.Config{
		Level:      logger.InfoLevel,
		Output:     &buf,
		TimeFormat: "2006-01-02",
	})
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	
	middleware := LoggingMiddleware(log)
	wrapped := middleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	
	wrapped.ServeHTTP(rec, req)
	
	output := buf.String()
	if !strings.Contains(output, "HTTP request") {
		t.Error("log output missing HTTP request message")
	}
	if !strings.Contains(output, "GET") {
		t.Error("log output missing method")
	}
	if !strings.Contains(output, "/test") {
		t.Error("log output missing path")
	}
	if !strings.Contains(output, "200") {
		t.Error("log output missing status code")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&logger.Config{
		Level:      logger.InfoLevel,
		Output:     &buf,
		TimeFormat: "2006-01-02",
	})
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	
	middleware := RecoveryMiddleware(log)
	wrapped := middleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	
	wrapped.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	
	output := buf.String()
	if !strings.Contains(output, "panic recovered") {
		t.Error("log output missing panic recovered message")
	}
	if !strings.Contains(output, "test panic") {
		t.Error("log output missing panic message")
	}
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		expectHeader   bool
		method         string
		expectStatus   int
	}{
		{
			name:           "Allow all origins",
			allowedOrigins: []string{"*"},
			origin:         "http://example.com",
			expectHeader:   true,
			method:         "GET",
			expectStatus:   http.StatusOK,
		},
		{
			name:           "Allow specific origin",
			allowedOrigins: []string{"http://example.com"},
			origin:         "http://example.com",
			expectHeader:   true,
			method:         "GET",
			expectStatus:   http.StatusOK,
		},
		{
			name:           "Deny unallowed origin",
			allowedOrigins: []string{"http://example.com"},
			origin:         "http://evil.com",
			expectHeader:   false,
			method:         "GET",
			expectStatus:   http.StatusOK,
		},
		{
			name:           "Handle preflight",
			allowedOrigins: []string{"*"},
			origin:         "http://example.com",
			expectHeader:   true,
			method:         "OPTIONS",
			expectStatus:   http.StatusNoContent,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			middleware := CORSMiddleware(tt.allowedOrigins)
			wrapped := middleware(handler)
			
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			
			wrapped.ServeHTTP(rec, req)
			
			if rec.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rec.Code)
			}
			
			corsHeader := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.expectHeader && corsHeader == "" {
				t.Error("expected CORS header but got none")
			}
			if !tt.expectHeader && corsHeader != "" {
				t.Errorf("expected no CORS header but got %s", corsHeader)
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value("request_id")
		if reqID == nil {
			t.Error("request ID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := RequestIDMiddleware()
	wrapped := middleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	
	wrapped.ServeHTTP(rec, req)
	
	reqIDHeader := rec.Header().Get("X-Request-ID")
	if reqIDHeader == "" {
		t.Error("X-Request-ID header not set")
	}
}