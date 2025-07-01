package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ayutaz/orochi/internal/logger"
)

func TestDatabase(t *testing.T) {
	// Create temporary database
	tmpDir, err := os.MkdirTemp("", "orochi-db-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	log := logger.NewWithLevel(logger.InfoLevel)

	db, err := NewDB(dbPath, log)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Test torrent operations
	t.Run("TorrentOperations", func(t *testing.T) {
		record := &TorrentRecord{
			ID:           "test-id",
			InfoHash:     "1234567890abcdef1234567890abcdef12345678",
			Name:         "Test Torrent",
			Size:         1024 * 1024 * 100, // 100MB
			Status:       "downloading",
			Progress:     50.5,
			Downloaded:   1024 * 1024 * 50,
			Uploaded:     1024 * 1024 * 10,
			DownloadPath: "/tmp/downloads",
			AddedAt:      time.Now(),
			Metadata:     `{"announce":"http://tracker.example.com"}`,
		}

		// Save torrent
		if err := db.SaveTorrent(record); err != nil {
			t.Errorf("failed to save torrent: %v", err)
		}

		// Get torrent
		retrieved, err := db.GetTorrent("test-id")
		if err != nil {
			t.Errorf("failed to get torrent: %v", err)
		}

		if retrieved.Name != record.Name {
			t.Errorf("expected name %s, got %s", record.Name, retrieved.Name)
		}

		// Update progress
		if err := db.UpdateTorrentProgress("test-id", 75.0, 1024*1024*75, 1024*1024*20); err != nil {
			t.Errorf("failed to update progress: %v", err)
		}

		// List torrents
		torrents, err := db.ListTorrents()
		if err != nil {
			t.Errorf("failed to list torrents: %v", err)
		}

		if len(torrents) != 1 {
			t.Errorf("expected 1 torrent, got %d", len(torrents))
		}

		// Mark completed
		if err := db.MarkTorrentCompleted("test-id"); err != nil {
			t.Errorf("failed to mark completed: %v", err)
		}

		// Delete torrent
		if err := db.DeleteTorrent("test-id"); err != nil {
			t.Errorf("failed to delete torrent: %v", err)
		}

		// Verify deleted
		_, err = db.GetTorrent("test-id")
		if err == nil {
			t.Error("expected error for deleted torrent")
		}
	})

	// Test settings operations
	t.Run("SettingsOperations", func(t *testing.T) {
		// Save setting
		if err := db.SaveSetting("download_path", "/home/user/downloads"); err != nil {
			t.Errorf("failed to save setting: %v", err)
		}

		// Get setting
		value, err := db.GetSetting("download_path")
		if err != nil {
			t.Errorf("failed to get setting: %v", err)
		}

		if value != "/home/user/downloads" {
			t.Errorf("expected /home/user/downloads, got %s", value)
		}

		// Save JSON settings
		settings := map[string]interface{}{
			"language":         "ja",
			"theme":            "dark",
			"maxConnections":   200,
			"downloadPath":     "/downloads",
			"maxDownloadSpeed": 0,
			"maxUploadSpeed":   0,
		}

		if err := db.SaveSettingsJSON(settings); err != nil {
			t.Errorf("failed to save JSON settings: %v", err)
		}

		// Get JSON settings
		var retrieved map[string]interface{}
		if err := db.GetSettingsJSON(&retrieved); err != nil {
			t.Errorf("failed to get JSON settings: %v", err)
		}

		if retrieved["language"] != "ja" {
			t.Errorf("expected language ja, got %v", retrieved["language"])
		}
	})
}
