package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "orochi-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	configPath := filepath.Join(tmpDir, "config.json")
	testConfig := &Config{
		Port:        9999,
		DownloadDir: "/custom/downloads",
		MaxTorrents: 20,
		MaxPeers:    500,
	}
	
	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create loader with custom path
	loader := &Loader{
		configPaths: []string{configPath},
		envPrefix:   "TEST_OROCHI_",
	}
	
	config, err := loader.Load()
	if err != nil {
		t.Fatal(err)
	}
	
	if config.Port != 9999 {
		t.Errorf("expected port 9999, got %d", config.Port)
	}
	if config.DownloadDir != "/custom/downloads" {
		t.Errorf("expected download dir /custom/downloads, got %s", config.DownloadDir)
	}
	if config.MaxTorrents != 20 {
		t.Errorf("expected max torrents 20, got %d", config.MaxTorrents)
	}
	if config.MaxPeers != 500 {
		t.Errorf("expected max peers 500, got %d", config.MaxPeers)
	}
}

func TestLoader_LoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_OROCHI_PORT", "7777")
	os.Setenv("TEST_OROCHI_DOWNLOAD_DIR", "/env/downloads")
	os.Setenv("TEST_OROCHI_MAX_TORRENTS", "30")
	os.Setenv("TEST_OROCHI_MAX_PEERS", "600")
	os.Setenv("TEST_OROCHI_VPN_INTERFACE", "tun0")
	
	defer func() {
		os.Unsetenv("TEST_OROCHI_PORT")
		os.Unsetenv("TEST_OROCHI_DOWNLOAD_DIR")
		os.Unsetenv("TEST_OROCHI_MAX_TORRENTS")
		os.Unsetenv("TEST_OROCHI_MAX_PEERS")
		os.Unsetenv("TEST_OROCHI_VPN_INTERFACE")
	}()
	
	loader := &Loader{
		configPaths: []string{}, // No config files
		envPrefix:   "TEST_OROCHI_",
	}
	
	config, err := loader.Load()
	if err != nil {
		t.Fatal(err)
	}
	
	if config.Port != 7777 {
		t.Errorf("expected port 7777, got %d", config.Port)
	}
	if config.DownloadDir != "/env/downloads" {
		t.Errorf("expected download dir /env/downloads, got %s", config.DownloadDir)
	}
	if config.MaxTorrents != 30 {
		t.Errorf("expected max torrents 30, got %d", config.MaxTorrents)
	}
	if config.MaxPeers != 600 {
		t.Errorf("expected max peers 600, got %d", config.MaxPeers)
	}
	if config.VPNInterface != "tun0" {
		t.Errorf("expected VPN interface tun0, got %s", config.VPNInterface)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "orochi-save-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	config := &Config{
		Port:         8888,
		DownloadDir:  "/saved/downloads",
		MaxTorrents:  15,
		MaxPeers:     300,
		VPNInterface: "vpn0",
	}
	
	configPath := filepath.Join(tmpDir, "subdir", "config.json")
	
	if err := SaveConfig(config, configPath); err != nil {
		t.Fatal(err)
	}
	
	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Error("config file not created")
	}
	
	// Load and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	
	if loaded.Port != config.Port {
		t.Errorf("saved port mismatch: expected %d, got %d", config.Port, loaded.Port)
	}
	if loaded.DownloadDir != config.DownloadDir {
		t.Errorf("saved download dir mismatch: expected %s, got %s", config.DownloadDir, loaded.DownloadDir)
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"8080", 8080},
		{" 9090 ", 9090},
		{"invalid", 0},
		{"", 0},
		{"65535", 65535},
		{"99999", 99999}, // Will be caught by validation
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePort(tt.input)
			if result != tt.expected {
				t.Errorf("parsePort(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{" 100 ", 100},
		{"invalid", 0},
		{"", 0},
		{"-5", -5},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}