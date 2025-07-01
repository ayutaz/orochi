package torrent

import (
	"os"
	"testing"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
)

func TestClientAdapter(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-adapter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port
		DownloadDir: tmpDir,
		DataDir:     tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Test AddTorrent
	torrentData := CreateTestTorrent()
	id, err := adapter.AddTorrent(torrentData)
	if err != nil {
		t.Fatalf("failed to add torrent: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty ID")
	}

	// Test GetTorrent
	torr, ok := adapter.GetTorrent(id)
	if !ok {
		t.Fatal("failed to get torrent")
	}

	if torr.ID != id {
		t.Errorf("ID mismatch: got %s, want %s", torr.ID, id)
	}

	if torr.Info == nil || torr.Info.Name == "" {
		t.Error("expected non-empty name")
	}

	// Test ListTorrents
	torrents := adapter.ListTorrents()
	if len(torrents) != 1 {
		t.Errorf("expected 1 torrent, got %d", len(torrents))
	}

	// Test Count
	count := adapter.Count()
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Test StartTorrent
	err = adapter.StartTorrent(id)
	if err != nil {
		t.Errorf("failed to start torrent: %v", err)
	}

	// Test StopTorrent
	err = adapter.StopTorrent(id)
	if err != nil {
		t.Errorf("failed to stop torrent: %v", err)
	}

	// Test RemoveTorrent
	err = adapter.RemoveTorrent(id)
	if err != nil {
		t.Errorf("failed to remove torrent: %v", err)
	}

	// Verify removal
	count = adapter.Count()
	if count != 0 {
		t.Errorf("expected count 0 after removal, got %d", count)
	}
}

func TestClientAdapterErrors(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-adapter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port
		DownloadDir: tmpDir,
		DataDir:     tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Test GetTorrent with non-existent ID
	_, ok := adapter.GetTorrent("non-existent")
	if ok {
		t.Error("expected false for non-existent torrent")
	}

	// Test RemoveTorrent with non-existent ID
	err = adapter.RemoveTorrent("non-existent")
	if err == nil {
		t.Error("expected error for non-existent torrent")
	}

	// Test StartTorrent with non-existent ID
	err = adapter.StartTorrent("non-existent")
	if err == nil {
		t.Error("expected error for non-existent torrent")
	}

	// Test StopTorrent with non-existent ID
	err = adapter.StopTorrent("non-existent")
	if err == nil {
		t.Error("expected error for non-existent torrent")
	}

}

func TestClientAdapterMapStatus(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-adapter-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port
		DownloadDir: tmpDir,
		DataDir:     tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	tests := []struct {
		input    string
		expected Status
	}{
		{"downloading", StatusDownloading},
		{"seeding", StatusSeeding},
		{"stopped", StatusStopped},
		{"unknown", StatusStopped},
	}

	for _, tt := range tests {
		result := adapter.mapStatus(tt.input)
		if result != tt.expected {
			t.Errorf("mapStatus(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
