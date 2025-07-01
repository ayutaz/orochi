package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ayutaz/orochi/internal/config"
)

// Server represents the HTTP server.
type Server struct {
	config     *config.Config
	httpServer *http.Server
	mux        *http.ServeMux
}

// NewServer creates a new HTTP server.
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}

	// Set up routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/", s.handleHome)
}

// ServeHTTP implements http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// handleHome handles the home page.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	// Strict path matching - only serve home page for exact "/" path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	// For now, just return a simple message
	// Later this will serve the actual UI
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Orochi - Simple Torrent Client</title>
</head>
<body>
    <h1>Orochi</h1>
    <p>Simple Torrent Client</p>
</body>
</html>`)
}