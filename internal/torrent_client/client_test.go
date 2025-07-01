package torrentclient

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/zeebo/bencode"
)

// createTestTorrent creates a simple torrent file for testing.
func createTestTorrent() []byte {
	return createTestTorrentWithName("test.txt")
}

// createTestTorrentWithName creates a simple torrent file with a specific name.
func createTestTorrentWithName(name string) []byte {
	type bencodeInfo struct {
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Length      int    `bencode:"length"`
		Pieces      string `bencode:"pieces"`
	}

	type bencodeTorrent struct {
		Announce string      `bencode:"announce"`
		Info     bencodeInfo `bencode:"info"`
	}

	torrent := bencodeTorrent{
		Announce: "http://example.com:8000",
		Info: bencodeInfo{
			Name:        name,
			PieceLength: 16384,
			Length:      1024,
			Pieces:      "01234567890123456789",
		},
	}

	data, err := bencode.EncodeBytes(torrent)
	if err != nil {
		panic(err) // This should never happen in tests
	}

	return data
}

func TestNewClient(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestAddTorrent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port1,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create test torrent
	torrentData := createTestTorrent()
	ctx := context.Background()

	// Add torrent
	torr, err := client.AddTorrent(ctx, torrentData)
	if err != nil {
		t.Fatalf("failed to add torrent: %v", err)
	}

	// Check torrent properties
	if torr.InfoHash() == "" {
		t.Error("expected non-empty info hash")
	}

	if torr.Name() == "" {
		t.Error("expected non-empty name")
	}

	if torr.Length() <= 0 {
		t.Error("expected positive length")
	}
}

func TestAddMagnet(t *testing.T) {
	t.Skip("Skipping magnet test as it requires network access")

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port2,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Test magnet link (Ubuntu 20.04 LTS)
	magnetLink := "magnet:?xt=urn:btih:3c5b316981c3b7c6e2a774e5efe4c5763e48f277&dn=ubuntu-20.04.6-desktop-amd64.iso"
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Add magnet
	_, err = client.AddMagnet(ctx, magnetLink)
	if err != nil {
		t.Fatalf("failed to add magnet: %v", err)
	}
}

func TestGetTorrent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port3,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Add a torrent first
	torrentData := createTestTorrent()
	ctx := context.Background()
	torr, err := client.AddTorrent(ctx, torrentData)
	if err != nil {
		t.Fatalf("failed to add torrent: %v", err)
	}

	// Get torrent by info hash
	gotTorr, err := client.GetTorrent(torr.InfoHash())
	if err != nil {
		t.Fatalf("failed to get torrent: %v", err)
	}

	if gotTorr.InfoHash() != torr.InfoHash() {
		t.Errorf("info hash mismatch: got %s, want %s", gotTorr.InfoHash(), torr.InfoHash())
	}

	// Try to get non-existent torrent
	_, err = client.GetTorrent("0000000000000000000000000000000000000000")
	if err == nil {
		t.Error("expected error for non-existent torrent")
	}
}

func TestListTorrents(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port4,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Initially should be empty
	torrents := client.ListTorrents()
	if len(torrents) != 0 {
		t.Errorf("expected 0 torrents, got %d", len(torrents))
	}

	// Add some torrents
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		torrentData := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
		_, err := client.AddTorrent(ctx, torrentData)
		if err != nil {
			t.Fatalf("failed to add torrent %d: %v", i, err)
		}
	}

	// List should now have 3
	torrents = client.ListTorrents()
	if len(torrents) != 3 {
		t.Errorf("expected 3 torrents, got %d", len(torrents))
	}
}

func TestTorrentOperations(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port5,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Add torrent
	torrentData := createTestTorrent()
	ctx := context.Background()
	torr, err := client.AddTorrent(ctx, torrentData)
	if err != nil {
		t.Fatalf("failed to add torrent: %v", err)
	}

	// Test Start
	torr.Start()

	// Test Stop
	torr.Stop()

	// Test Progress
	progress := torr.Progress()
	if progress < 0 || progress > 100 {
		t.Errorf("invalid progress: %f", progress)
	}

	// Test Status
	status := torr.Status()
	if status == "" {
		t.Error("expected non-empty status")
	}

	// Test Files
	files := torr.Files()
	if len(files) == 0 {
		t.Error("expected at least one file")
	}

	// Test SavePath
	savePath := torr.SavePath()
	if !strings.Contains(savePath, tmpDir) {
		t.Errorf("save path should contain download dir: %s", savePath)
	}

	// Test Stats
	stats := torr.Stats()
	if stats.TotalPeers < 0 {
		t.Error("invalid total peers")
	}

	// Test Remove
	err = torr.Remove()
	if err != nil {
		t.Errorf("failed to remove torrent: %v", err)
	}

	// Verify it's removed
	torrents := client.ListTorrents()
	if len(torrents) != 0 {
		t.Error("torrent should be removed")
	}
}

func TestStatsString(t *testing.T) {
	stats := Stats{
		BytesReadData:    1024,
		BytesWrittenData: 2048,
		ActivePeers:      5,
		TotalPeers:       20,
		ConnectedSeeders: 3,
	}

	str := stats.String()
	expected := "Read: 1024 bytes, Written: 2048 bytes, Peers: 5/20, Seeders: 3"
	if str != expected {
		t.Errorf("expected %q, got %q", expected, str)
	}
}

func TestInvalidInfoHash(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "orochi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{
		Port:        0, // Use random port6,
		DownloadDir: tmpDir,
	}
	log := logger.NewWithLevel(logger.ErrorLevel)

	client, err := NewClient(cfg, log)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Test invalid info hash
	_, err = client.GetTorrent("invalid-hash")
	if err == nil {
		t.Error("expected error for invalid info hash")
	}
}
