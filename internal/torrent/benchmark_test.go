package torrent

import (
	"fmt"
	"testing"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
)

// BenchmarkListTorrents benchmarks the performance of listing torrents.
func BenchmarkListTorrents(b *testing.B) {
	cfg := &config.Config{
		Port:        0, // Random port for testing
		DataDir:     b.TempDir(),
		DownloadDir: b.TempDir(),
	}
	log := logger.NewNop()

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		b.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Add test torrents
	numTorrents := []int{10, 100, 1000}
	for _, n := range numTorrents {
		b.Run(fmt.Sprintf("torrents_%d", n), func(b *testing.B) {
			// Reset manager
			adapter, err := NewClientAdapter(cfg, log)
			if err != nil {
				b.Fatalf("failed to create adapter: %v", err)
			}
			defer adapter.Close()

			// Add n torrents
			for i := 0; i < n; i++ {
				data := createTestTorrent(fmt.Sprintf("test%d", i))
				_, err := adapter.AddTorrent(data)
				if err != nil {
					b.Fatalf("failed to add torrent: %v", err)
				}
			}

			// Benchmark listing
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				torrents := adapter.ListTorrents()
				if len(torrents) != n {
					b.Fatalf("expected %d torrents, got %d", n, len(torrents))
				}
			}
		})
	}
}

// BenchmarkGetTorrent benchmarks the performance of getting individual torrents.
func BenchmarkGetTorrent(b *testing.B) {
	cfg := &config.Config{
		Port:        0,
		DataDir:     b.TempDir(),
		DownloadDir: b.TempDir(),
	}
	log := logger.NewNop()

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		b.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Add a test torrent
	data := createTestTorrent("benchmark")
	id, err := adapter.AddTorrent(data)
	if err != nil {
		b.Fatalf("failed to add torrent: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		torrent, ok := adapter.GetTorrent(id)
		if !ok {
			b.Fatal("torrent not found")
		}
		if torrent.ID != id {
			b.Fatalf("expected id %s, got %s", id, torrent.ID)
		}
	}
}

// BenchmarkBatchUpdate benchmarks the performance of batch updates.
func BenchmarkBatchUpdate(b *testing.B) {
	cfg := &config.Config{
		Port:        0,
		DataDir:     b.TempDir(),
		DownloadDir: b.TempDir(),
	}
	log := logger.NewNop()

	adapter, err := NewClientAdapter(cfg, log)
	if err != nil {
		b.Fatalf("failed to create adapter: %v", err)
	}
	defer adapter.Close()

	// Add test torrents
	numTorrents := 100
	torrents := make([]string, 0, numTorrents)
	for i := 0; i < numTorrents; i++ {
		data := createTestTorrent(fmt.Sprintf("batch%d", i))
		id, err := adapter.AddTorrent(data)
		if err != nil {
			b.Fatalf("failed to add torrent: %v", err)
		}
		torrents = append(torrents, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate progress updates
		for j, id := range torrents {
			torrent, ok := adapter.GetTorrent(id)
			if !ok {
				b.Fatal("torrent not found")
			}
			// Update progress
			torrent.Progress = float64(j) / float64(numTorrents) * 100
			torrent.Downloaded = int64(j * 1024 * 1024)
			torrent.Uploaded = int64(j * 512 * 1024)
		}
	}
}