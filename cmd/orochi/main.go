package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/ayutaz/orochi/internal/torrent"
	"github.com/ayutaz/orochi/internal/web"
)

var (
	// Version information (set during build).
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		showVersion bool
		port        int
		downloadDir string
		useReal     bool
	)
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&downloadDir, "download-dir", "./downloads", "Download directory")
	flag.BoolVar(&useReal, "real", false, "Use real torrent client (experimental)")
	flag.Parse()

	if showVersion {
		fmt.Printf("Orochi %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Show disclaimer
	showDisclaimer()

	// Load configuration
	cfg := config.LoadDefault()
	if port != 8080 {
		cfg.Port = port
	}
	if downloadDir != "./downloads" {
		cfg.DownloadDir = downloadDir
	}

	// Create logger
	log := logger.NewWithLevel(logger.InfoLevel)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal("Invalid configuration", logger.Err(err))
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(cfg.GetAbsoluteDownloadDir(), 0o755); err != nil {
		log.Fatal("Failed to create download directory", logger.Err(err))
	}

	// Create torrent manager
	var manager torrent.Manager
	if useReal {
		log.Info("Using real BitTorrent client (experimental)")
		adapter, err := torrent.NewClientAdapter(cfg, log)
		if err != nil {
			log.Fatal("Failed to create torrent client", logger.Err(err))
		}
		manager = adapter
	} else {
		log.Info("Using stub torrent manager")
		manager = torrent.NewManager()
	}

	// Create and configure web server
	server := web.NewServer(cfg)
	server.SetTorrentManager(manager)

	// Start server in background
	go func() {
		log.Info("Starting Orochi", logger.Int("port", cfg.Port))
		if err := server.Start(); err != nil {
			log.Fatal("Server failed", logger.Err(err))
		}
	}()

	// Start progress updater
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			// Send actual torrent data via WebSocket
			server.BroadcastTorrentData()
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	log.Info("Shutting down...")
	if err := server.Shutdown(); err != nil {
		log.Error("Server shutdown error", logger.Err(err))
	}
}

func showDisclaimer() {
	fmt.Print(`
================================================================================
                               OROCHI - DISCLAIMER
================================================================================

This software is designed for downloading and sharing files using the BitTorrent
protocol. The use of this software for downloading or distributing copyrighted
material without permission is illegal.

Users are responsible for complying with local laws and regulations.

By using this software, you agree that:
- You will only use it for legal purposes
- You will not download or distribute copyrighted material without permission
- The developers are not responsible for any misuse of this software

================================================================================
`)
}
