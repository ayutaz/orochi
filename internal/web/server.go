package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/torrent"
)

// Server represents the HTTP server.
type Server struct {
	config         *config.Config
	httpServer     *http.Server
	mux            *http.ServeMux
	torrentManager *torrent.Manager
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

// SetTorrentManager sets the torrent manager for the server.
func (s *Server) SetTorrentManager(tm *torrent.Manager) {
	s.torrentManager = tm
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// Health check
	s.mux.HandleFunc("/health", s.handleHealth)
	
	// API routes
	s.mux.HandleFunc("/api/torrents", s.handleAPITorrents)
	s.mux.HandleFunc("/api/torrents/", s.handleAPITorrent)
	s.mux.HandleFunc("/api/torrents/magnet", s.handleAPITorrentMagnet)
	
	// Web UI
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
	// Serve static files
	if r.URL.Path != "/" {
		// Try to serve static files
		http.FileServer(http.Dir("web")).ServeHTTP(w, r)
		return
	}
	
	// Serve index.html for root path
	http.ServeFile(w, r, "web/templates/index.html")
}