package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader handles configuration loading from multiple sources.
type Loader struct {
	configPaths []string
	envPrefix   string
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		configPaths: []string{
			"orochi.json",
			"config.json",
			filepath.Join(getHomeDir(), ".orochi", "config.json"),
			"/etc/orochi/config.json",
		},
		envPrefix: "OROCHI_",
	}
}

// Load loads configuration from available sources.
func (l *Loader) Load() (*Config, error) {
	// Start with default config
	config := LoadDefault()

	// Try to load from config file
	l.loadFromFile(config) // Config file is optional

	// Override with environment variables
	l.loadFromEnv(config)

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromFile loads configuration from a JSON file.
func (l *Loader) loadFromFile(config *Config) {
	for _, path := range l.configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Try next path
		}

		if err := json.Unmarshal(data, config); err != nil {
			continue // Try next path
		}

		// Successfully loaded
		return
	}

	// No config file found (which is OK)
}

// loadFromEnv loads configuration from environment variables.
func (l *Loader) loadFromEnv(config *Config) {
	// Check for port
	if portStr := os.Getenv(l.envPrefix + "PORT"); portStr != "" {
		if port := parsePort(portStr); port > 0 {
			config.Port = port
		}
	}

	// Check for download directory
	if dir := os.Getenv(l.envPrefix + "DOWNLOAD_DIR"); dir != "" {
		config.DownloadDir = dir
	}

	// Check for max torrents
	if maxStr := os.Getenv(l.envPrefix + "MAX_TORRENTS"); maxStr != "" {
		if maxTorrents := parseInt(maxStr); maxTorrents > 0 {
			config.MaxTorrents = maxTorrents
		}
	}

	// Check for max peers
	if maxStr := os.Getenv(l.envPrefix + "MAX_PEERS"); maxStr != "" {
		if maxPeers := parseInt(maxStr); maxPeers > 0 {
			config.MaxPeers = maxPeers
		}
	}

	// Check for VPN interface
	if vpn := os.Getenv(l.envPrefix + "VPN_INTERFACE"); vpn != "" {
		config.VPNInterface = vpn
	}
}

// SaveConfig saves the configuration to a file.
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0o600)
}

// getHomeDir returns the user's home directory.
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// parsePort parses a port number from string.
func parsePort(s string) int {
	s = strings.TrimSpace(s)
	var port int
	_, _ = fmt.Sscanf(s, "%d", &port)
	return port
}

// parseInt parses an integer from string.
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	var num int
	_, _ = fmt.Sscanf(s, "%d", &num)
	return num
}
