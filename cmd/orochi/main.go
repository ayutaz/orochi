package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ayutaz/orochi/internal/config"
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
	)
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&downloadDir, "download-dir", "./downloads", "Download directory")
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(cfg.GetAbsoluteDownloadDir(), 0o755); err != nil {
		log.Fatalf("Failed to create download directory: %v", err)
	}

	// Create torrent manager
	manager := torrent.NewManager()

	// Create and configure web server
	server := web.NewServer(cfg)
	server.SetTorrentManager(manager)

	// Start server in background
	go func() {
		log.Printf("Starting Orochi on http://localhost:%d", cfg.Port)
		if err := server.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// Graceful shutdown
	log.Println("Shutting down...")
	if err := server.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
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
