package config

import (
	"errors"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	Port         int    `json:"port"`
	DownloadDir  string `json:"download_dir"`
	MaxTorrents  int    `json:"max_torrents"`
	MaxPeers     int    `json:"max_peers"`
	VPNInterface string `json:"vpn_interface,omitempty"`
}

// LoadDefault returns the default configuration.
func LoadDefault() *Config {
	return &Config{
		Port:        8080,
		DownloadDir: "./downloads",
		MaxTorrents: 5,
		MaxPeers:    200,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("invalid port number: must be between 1 and 65535")
	}
	
	if c.DownloadDir == "" {
		return errors.New("download directory cannot be empty")
	}
	
	if c.MaxTorrents < 1 {
		return errors.New("max torrents must be at least 1")
	}
	
	if c.MaxPeers < 1 {
		return errors.New("max peers must be at least 1")
	}
	
	return nil
}

// GetAbsoluteDownloadDir returns the absolute path of the download directory.
func (c *Config) GetAbsoluteDownloadDir() string {
	if filepath.IsAbs(c.DownloadDir) {
		return c.DownloadDir
	}
	
	absPath, err := filepath.Abs(c.DownloadDir)
	if err != nil {
		// If we can't get absolute path, return the original
		return c.DownloadDir
	}
	
	return absPath
}