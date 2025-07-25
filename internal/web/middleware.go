package web

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ayutaz/orochi/internal/auth"
	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/ayutaz/orochi/internal/metrics"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for request ID.
	requestIDKey contextKey = "request_id"
)

// Middleware is a function that wraps an HTTP handler.
type Middleware func(http.Handler) http.Handler

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			log.Info("HTTP request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", wrapped.statusCode),
				logger.Duration("duration", duration),
				logger.String("remote_addr", r.RemoteAddr),
			)

			// Update metrics
			m := metrics.Get()
			m.IncrementHTTPRequests()
			m.RecordHTTPDuration(duration)
			if wrapped.statusCode >= 400 {
				m.IncrementHTTPErrors()
			}
		})
	}
}

// RecoveryMiddleware recovers from panics.
func RecoveryMiddleware(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Convert panic value to string
					var errStr string
					switch v := err.(type) {
					case string:
						errStr = v
					case error:
						errStr = v.Error()
					default:
						errStr = "panic occurred"
					}

					log.Error("panic recovered",
						logger.String("error", errStr),
						logger.String("method", r.Method),
						logger.String("path", r.URL.Path),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers.
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			// Check if origin is allowed
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware adds a request ID to the context.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple request ID generation
			requestID := generateRequestID()

			// Add to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, requestIDKey, requestID)

			// Add to response header
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireTorrentManager ensures the torrent manager is initialized.
func RequireTorrentManager(s *Server) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if s.torrentManager == nil {
				writeError(w, http.StatusInternalServerError, "torrent manager not initialized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// generateRequestID generates a simple request ID.
func generateRequestID() string {
	// Simple implementation - in production, use UUID
	return time.Now().Format("20060102150405") + "-" + generateRandomString(8)
}

// generateRandomString generates a random string of given length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		// Simple pseudo-random for demo - in production use crypto/rand
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// AuthMiddleware checks for valid authentication.
func AuthMiddleware(authManager interface{ ValidateToken(string) (string, error) }) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for certain paths
			path := r.URL.Path
			if path == "/health" || path == "/api/auth/login" || strings.HasPrefix(path, "/ws") {
				next.ServeHTTP(w, r)
				return
			}

			// Check for Bearer token
			authHeader := r.Header.Get("Authorization")
			token, ok := auth.ParseBearerToken(authHeader)
			if !ok {
				w.Header().Set("WWW-Authenticate", "Bearer")
				writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			// Validate token
			username, err := authManager.ValidateToken(token)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Bearer")
				if errors.IsAuthenticationFailed(err) {
					writeError(w, http.StatusUnauthorized, err.Error())
				} else {
					writeError(w, http.StatusInternalServerError, "authentication error")
				}
				return
			}

			// Add username to context
			ctx := context.WithValue(r.Context(), contextKey("username"), username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
