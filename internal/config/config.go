package config

import (
	"errors"
	"path/filepath"

	"github.com/ayutaz/orochi/internal/network"
)

// Validation errors.
var (
	ErrInvalidPort        = errors.New("invalid port number: must be between 1 and 65535")
	ErrEmptyDownloadDir   = errors.New("download directory cannot be empty")
	ErrInvalidMaxTorrents = errors.New("max torrents must be at least 1")
	ErrInvalidMaxPeers    = errors.New("max peers must be at least 1")
)

// Config represents the application configuration.
type Config struct {
	Port           int                  `json:"port"`
	DownloadDir    string               `json:"download_dir"`
	MaxTorrents    int                  `json:"max_torrents"`
	MaxPeers       int                  `json:"max_peers"`
	VPNInterface   string               `json:"vpn_interface,omitempty"` // Deprecated: use VPN.InterfaceName
	DataDir        string               `json:"data_dir,omitempty"`
	AllowedOrigins []string             `json:"allowed_origins,omitempty"`
	VPN            *network.VPNConfig   `json:"vpn,omitempty"`
}

// LoadDefault returns the default configuration.
func LoadDefault() *Config {
	return &Config{
		Port:           8080,
		DownloadDir:    "./downloads",
		MaxTorrents:    5,
		MaxPeers:       200,
		DataDir:        "./data",
		AllowedOrigins: []string{}, // Empty means allow all origins
		VPN:            network.NewVPNConfig(),
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidPort
	}

	if c.DownloadDir == "" {
		return ErrEmptyDownloadDir
	}

	if c.MaxTorrents < 1 {
		return ErrInvalidMaxTorrents
	}

	if c.MaxPeers < 1 {
		return ErrInvalidMaxPeers
	}

	// Validate VPN config if present
	if c.VPN != nil {
		if err := c.VPN.Validate(); err != nil {
			return err
		}
	}

	// Handle deprecated VPNInterface field
	if c.VPNInterface != "" && c.VPN == nil {
		c.VPN = &network.VPNConfig{
			Enabled:       true,
			InterfaceName: c.VPNInterface,
			KillSwitch:    true,
		}
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
