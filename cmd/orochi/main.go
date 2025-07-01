package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	// Version information (set during build)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("Orochi %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// TODO: アプリケーションのメインロジックを実装
	fmt.Println("Orochi - Simple Torrent Client")
}