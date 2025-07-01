package torrent

import (
	"fmt"
	"sync"
	"testing"

	"github.com/zeebo/bencode"
)

// BenchmarkManagerAdd benchmarks adding torrents.
func BenchmarkManagerAdd(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create unique torrent data for each iteration
		uniqueData := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
		_, _ = manager.AddTorrent(uniqueData)
	}
}

// BenchmarkManagerGet benchmarks getting torrents.
func BenchmarkManagerGet(b *testing.B) {
	manager := NewManager()

	// Pre-populate with torrents
	var ids []string
	for i := 0; i < 1000; i++ {
		data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
		id, err := manager.AddTorrent(data)
		if err != nil {
			b.Fatalf("failed to add torrent: %v", err)
		}
		ids = append(ids, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetTorrent(ids[i%len(ids)])
	}
}

// BenchmarkManagerList benchmarks listing torrents.
func BenchmarkManagerList(b *testing.B) {
	manager := NewManager()

	// Pre-populate with torrents
	for i := 0; i < 100; i++ {
		data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
		_, _ = manager.AddTorrent(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ListTorrents()
	}
}

// BenchmarkManagerConcurrentOperations benchmarks concurrent access.
func BenchmarkManagerConcurrentOperations(b *testing.B) {
	manager := NewManager()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				// Add
				data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
				_, _ = manager.AddTorrent(data)
			case 1:
				// Get
				manager.GetTorrent("somekey")
			case 2:
				// List
				manager.ListTorrents()
			case 3:
				// Count
				manager.Count()
			}
			i++
		}
	})
}

// BenchmarkManagerWithRWMutex benchmarks the current RWMutex implementation.
func BenchmarkManagerWithRWMutex(b *testing.B) {
	manager := NewManager()

	// Pre-populate
	for i := 0; i < 100; i++ {
		data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
		_, _ = manager.AddTorrent(data)
	}

	b.Run("90%Read-10%Write", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					// Write operation
					data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
					_, _ = manager.AddTorrent(data)
				} else {
					// Read operation
					manager.Count()
				}
				i++
			}
		})
	})

	b.Run("50%Read-50%Write", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					// Write operation
					data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
					_, _ = manager.AddTorrent(data)
				} else {
					// Read operation
					manager.ListTorrents()
				}
				i++
			}
		})
	})
}

// Helper function to create test torrent with specific name.
func createTestTorrentWithName(name string) []byte {
	torrent := map[string]interface{}{
		"announce": "http://tracker.example.com:8080/announce",
		"info": map[string]interface{}{
			"name":         name,
			"piece length": int64(16384),
			"length":       int64(1024),
			"pieces":       string(make([]byte, 20)), // Single piece
		},
	}

	data, err := bencode.EncodeBytes(torrent)
	if err != nil {
		return nil
	}
	return data
}

// BenchmarkConcurrentAccess tests concurrent performance.
func BenchmarkConcurrentAccess(b *testing.B) {
	benchmarks := []struct {
		name    string
		readers int
		writers int
	}{
		{"1Reader-1Writer", 1, 1},
		{"10Readers-1Writer", 10, 1},
		{"10Readers-5Writers", 10, 5},
		{"100Readers-10Writers", 100, 10},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			manager := NewManager()

			// Pre-populate
			for i := 0; i < 100; i++ {
				data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
				_, _ = manager.AddTorrent(data)
			}

			b.ResetTimer()

			var wg sync.WaitGroup
			stop := make(chan struct{})

			// Start readers
			for i := 0; i < bm.readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-stop:
							return
						default:
							manager.Count()
							manager.ListTorrents()
						}
					}
				}()
			}

			// Start writers
			for i := 0; i < bm.writers; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					j := 0
					for {
						select {
						case <-stop:
							return
						default:
							data := createTestTorrentWithName(fmt.Sprintf("writer%d-test%d.txt", id, j))
							_, _ = manager.AddTorrent(data)
							j++
						}
					}
				}(i)
			}

			// Run for the duration of the benchmark
			b.StopTimer()
			close(stop)
			wg.Wait()
		})
	}
}
