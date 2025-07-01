package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/ayutaz/orochi/internal/metrics"
	"github.com/ayutaz/orochi/internal/torrent"
)

// Server represents the HTTP server.
type Server struct {
	config         *config.Config
	httpServer     *http.Server
	router         *Router
	torrentManager torrent.Manager
	logger         logger.Logger
	wsHub          *Hub
}

// Router returns the server's router for testing.
func (s *Server) Router() http.Handler {
	return s.router
}

// NewServer creates a new HTTP server.
func NewServer(cfg *config.Config) *Server {
	// Create logger
	log := logger.NewWithLevel(logger.InfoLevel).WithFields(
		logger.String("component", "web-server"),
	)

	s := &Server{
		config: cfg,
		router: NewRouter(),
		logger: log,
		wsHub:  NewHub(log, cfg.AllowedOrigins),
	}

	// Set up middleware
	s.router.Use(
		RequestIDMiddleware(),
		LoggingMiddleware(log),
		RecoveryMiddleware(log),
		CORSMiddleware([]string{"*"}),
	)

	// Set up routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// SetTorrentManager sets the torrent manager for the server.
func (s *Server) SetTorrentManager(tm torrent.Manager) {
	s.torrentManager = tm
}

// BroadcastTorrentUpdate sends torrent update to all connected clients.
func (s *Server) BroadcastTorrentUpdate() {
	if s.wsHub != nil {
		s.wsHub.BroadcastTorrentUpdate()
	}
}

// BroadcastTorrentData sends actual torrent data to all connected clients.
func (s *Server) BroadcastTorrentData() {
	if s.wsHub != nil && s.torrentManager != nil {
		torrents := s.torrentManager.ListTorrents()

		// Convert to API response format
		torrentResponses := make([]TorrentResponse, 0, len(torrents))
		for _, t := range torrents {
			torrentResponses = append(torrentResponses, toTorrentResponse(t))
		}

		s.wsHub.BroadcastTorrentData(torrentResponses)
	}
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.wrapHandler(s.handleHealth))

	// Metrics endpoint
	s.router.GET("/metrics", s.wrapHandler(s.handleMetrics))

	// API routes group
	api := s.router.Group("/api", RequireTorrentManager(s))

	// Torrent endpoints
	api.GET("/torrents", s.wrapHandler(s.handleListTorrents))
	api.POST("/torrents", s.wrapHandler(s.handleAddTorrent))
	api.POST("/torrents/magnet", s.wrapHandler(s.handleAddMagnet))
	api.GET("/torrents/:id", s.wrapHandler(s.handleGetTorrent))
	api.DELETE("/torrents/:id", s.wrapHandler(s.handleDeleteTorrent))
	api.POST("/torrents/:id/start", s.wrapHandler(s.handleStartTorrent))
	api.POST("/torrents/:id/stop", s.wrapHandler(s.handleStopTorrent))
	api.PUT("/torrents/:id/files", s.wrapHandler(s.handleUpdateFiles))

	// Settings endpoints
	api.GET("/settings", s.wrapHandler(s.handleGetSettings))
	api.PUT("/settings", s.wrapHandler(s.handleUpdateSettings))

	// VPN endpoints
	api.GET("/vpn/status", s.wrapHandler(s.handleGetVPNStatus))
	api.PUT("/vpn/config", s.wrapHandler(s.handleUpdateVPNConfig))

	// WebSocket endpoint
	s.router.GET("/ws", s.wrapHandler(s.handleWebSocket))

	// Web UI
	s.router.GET("/", s.wrapHandler(s.handleHome))
	s.router.GET("/*", s.wrapHandler(s.handleStatic))
}

// wrapHandler converts a standard handler to our HandlerFunc.
func (s *Server) wrapHandler(handler func(w http.ResponseWriter, r *http.Request)) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		handler(w, r)
		return nil
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// handleHome handles the home page.
func (s *Server) handleHome(w http.ResponseWriter, _ *http.Request) {
	// Try to serve React app from dist directory first
	staticFS, err := GetStaticFS()
	if err == nil {
		indexHTML, err := fs.ReadFile(staticFS, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, err := w.Write(indexHTML); err != nil {
				s.logger.Error("failed to write response", logger.Err(err))
			}
			return
		}
	}

	// Fallback to templates directory
	templatesFS, err := GetTemplatesFS()
	if err != nil {
		s.logger.Error("failed to get templates FS", logger.Err(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	indexHTML, err := fs.ReadFile(templatesFS, "index.html")
	if err != nil {
		s.logger.Error("failed to read index.html", logger.Err(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(indexHTML); err != nil {
		s.logger.Error("failed to write response", logger.Err(err))
	}
}

// handleStatic handles static file requests.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	staticFS, err := GetStaticFS()
	if err != nil {
		s.logger.Error("failed to get static FS", logger.Err(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// For React app routes, serve index.html
	path := r.URL.Path
	if path == "/torrent" || path == "/settings" ||
		(len(path) > 8 && path[:8] == "/torrent/") {
		indexHTML, err := fs.ReadFile(staticFS, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexHTML)
			return
		}
	}

	http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
}

// handleMetrics handles metrics endpoint.
func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	// Update system metrics
	m := metrics.Get()

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	if memStats.Alloc <= 9223372036854775807 { // max int64
		m.SetMemoryUsage(int64(memStats.Alloc))
	}

	// Get goroutine count
	numGoroutines := runtime.NumGoroutine()
	if numGoroutines <= 2147483647 { // max int32
		m.SetGoroutineCount(int32(numGoroutines)) //nolint:gosec // bounds checked above
	}

	// Get torrent metrics if manager is available
	if s.torrentManager != nil {
		torrents := s.torrentManager.ListTorrents()
		m.TorrentsTotal = int64(len(torrents))

		// Reset status counters
		m.TorrentsDownloading = 0
		m.TorrentsSeeding = 0
		m.TorrentsStopped = 0
		m.TorrentsError = 0

		// Count by status
		for _, t := range torrents {
			switch t.Status {
			case torrent.StatusDownloading:
				m.TorrentsDownloading++
			case torrent.StatusSeeding:
				m.TorrentsSeeding++
			case torrent.StatusStopped:
				m.TorrentsStopped++
			case torrent.StatusError:
				m.TorrentsError++
			}
		}
	}

	// Return metrics snapshot
	snapshot := m.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		s.logger.Error("failed to encode metrics", logger.Err(err))
	}
}

// saveConfig saves the current configuration to disk.
func (s *Server) saveConfig() error {
	// Get config file path
	configPath := filepath.Join(s.config.DataDir, "config.json")

	// Ensure data directory exists
	if err := os.MkdirAll(s.config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
