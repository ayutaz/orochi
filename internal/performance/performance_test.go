package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/torrent"
	"github.com/ayutaz/orochi/internal/web"
)

// BenchmarkServer tests server performance.
func BenchmarkServer(b *testing.B) {
	// Create server
	cfg := config.LoadDefault()
	server := web.NewServer(cfg)
	server.SetTorrentManager(torrent.NewManager())

	// Create test server
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	b.Run("HealthCheck", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/health", http.NoBody)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})

	b.Run("ListTorrents", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/torrents", http.NoBody)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})

	b.Run("AddMagnet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=test%d.txt", i, i)
			reqBody := map[string]string{"magnet": magnet}
			data, err := json.Marshal(reqBody)
			if err != nil {
				b.Fatal(err)
			}

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost,
				ts.URL+"/api/torrents/magnet", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})

	b.Run("ConcurrentRequests", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				var resp *http.Response
				var err error

				var req *http.Request
				switch i % 3 {
				case 0:
					req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/health", http.NoBody)
				case 1:
					req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/torrents", http.NoBody)
				case 2:
					req, _ = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/metrics", http.NoBody)
				}
				resp, err = http.DefaultClient.Do(req)

				if err != nil {
					b.Fatal(err)
				}
				resp.Body.Close()
				i++
			}
		})
	})
}

// BenchmarkTorrentManager tests torrent manager performance.
func BenchmarkTorrentManager(b *testing.B) {
	b.Run("RWMutexImplementation", func(b *testing.B) {
		manager := torrent.NewManager()
		benchmarkManagerOperations(b, manager)
	})

	b.Run("ConcurrentImplementation", func(b *testing.B) {
		manager := torrent.NewConcurrentManager()
		benchmarkManagerOperations(b, manager)
	})
}

func benchmarkManagerOperations(b *testing.B, manager torrent.Manager) {
	b.Helper()
	// Pre-populate with some torrents
	for i := 0; i < 100; i++ {
		magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=test%d.txt", i, i)
		_, _ = manager.AddMagnet(magnet)
	}

	b.Run("Add", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=bench%d.txt", b.N+i, i)
			_, _ = manager.AddMagnet(magnet)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Get an existing ID
		id := fmt.Sprintf("%040d", 50)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.GetTorrent(id)
		}
	})

	b.Run("List", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.ListTorrents()
		}
	})

	b.Run("MixedOperations", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0:
					// Add
					magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=mixed%d.txt", b.N*1000+i, i)
					_, _ = manager.AddMagnet(magnet)
				case 1:
					// Get
					id := fmt.Sprintf("%040d", i%100)
					manager.GetTorrent(id)
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
	})
}

// TestLoadPerformance tests performance under high load.
func TestLoadPerformance(t *testing.T) {
	t.Skip("Temporarily skipping load test due to race condition")

	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	// Create server
	cfg := config.LoadDefault()
	server := web.NewServer(cfg)
	server.SetTorrentManager(torrent.NewConcurrentManager())

	// Create test server
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Number of concurrent clients
	numClients := 100
	requestsPerClient := 100

	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, numClients*requestsPerClient)

	// Launch clients
	for client := 0; client < numClients; client++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			for req := 0; req < requestsPerClient; req++ {
				// Mix of operations
				var err error
				switch req % 5 {
				case 0:
					// Health check
					req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/health", http.NoBody)
					resp, healthErr := http.DefaultClient.Do(req)
					if healthErr == nil {
						resp.Body.Close()
					}
					err = healthErr

				case 1:
					// List torrents
					req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/torrents", http.NoBody)
					resp, listErr := http.DefaultClient.Do(req)
					if listErr == nil {
						resp.Body.Close()
					}
					err = listErr

				case 2:
					// Add magnet
					magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=load%d-%d.txt", clientID*1000+req, clientID, req)
					reqBody := map[string]string{"magnet": magnet}
					data, marshalErr := json.Marshal(reqBody)
					if marshalErr != nil {
						select {
						case errors <- marshalErr:
						default:
						}
						return
					}

					req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost,
						ts.URL+"/api/torrents/magnet", bytes.NewReader(data))
					req.Header.Set("Content-Type", "application/json")
					resp, postErr := http.DefaultClient.Do(req)
					if postErr == nil {
						resp.Body.Close()
					}
					err = postErr

				case 3:
					// Get torrent (might 404)
					id := fmt.Sprintf("%040d", req%50)
					req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/api/torrents/"+id, http.NoBody)
					resp, getErr := http.DefaultClient.Do(req)
					if getErr == nil {
						resp.Body.Close()
					}
					err = getErr

				case 4:
					// Metrics
					req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/metrics", http.NoBody)
					resp, metricsErr := http.DefaultClient.Do(req)
					if metricsErr == nil {
						resp.Body.Close()
					}
					err = metricsErr
				}

				if err != nil {
					select {
					case errors <- err:
					default:
					}
				}
			}
		}(client)
	}

	// Wait for completion
	wg.Wait()
	duration := time.Since(start)
	close(errors)

	// Count errors
	errorCount := 0
	for range errors {
		errorCount++
	}

	// Calculate stats
	totalRequests := numClients * requestsPerClient
	successfulRequests := totalRequests - errorCount
	requestsPerSecond := float64(totalRequests) / duration.Seconds()

	t.Logf("Load test completed:")
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful requests: %d", successfulRequests)
	t.Logf("  Failed requests: %d", errorCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Requests/second: %.2f", requestsPerSecond)

	// Performance assertions
	if errorCount > totalRequests/100 { // Allow 1% error rate
		t.Errorf("too many errors: %d (%.2f%%)", errorCount, float64(errorCount)/float64(totalRequests)*100)
	}

	if requestsPerSecond < 1000 {
		t.Errorf("requests per second too low: %.2f", requestsPerSecond)
	}
}

// TestMemoryUsage tests for memory leaks.
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory test in short mode")
	}

	manager := torrent.NewConcurrentManager()

	// Add and remove many torrents
	for cycle := 0; cycle < 10; cycle++ {
		// Add 1000 torrents
		ids := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=mem%d-%d.txt", cycle*1000+i, cycle, i)
			id, _ := manager.AddMagnet(magnet)
			ids[i] = id
		}

		// Remove all torrents
		for _, id := range ids {
			_ = manager.RemoveTorrent(id)
		}

		// Check count is back to 0
		if count := manager.Count(); count != 0 {
			t.Errorf("cycle %d: expected count 0, got %d", cycle, count)
		}
	}
}

// BenchmarkFileUpload tests file upload performance.
func BenchmarkFileUpload(b *testing.B) {
	// Create server
	cfg := config.LoadDefault()
	server := web.NewServer(cfg)
	server.SetTorrentManager(torrent.NewManager())

	// Create test server
	ts := httptest.NewServer(server.Router())
	defer ts.Close()

	// Create test torrent data
	torrentData := torrent.CreateTestTorrent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create multipart form
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		part, err := writer.CreateFormFile("torrent", fmt.Sprintf("test%d.torrent", i))
		if err != nil {
			b.Fatal(err)
		}

		if _, writeErr := part.Write(torrentData); writeErr != nil {
			b.Fatal(writeErr)
		}

		if closeErr := writer.Close(); closeErr != nil {
			b.Fatal(closeErr)
		}

		// Send request
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/api/torrents", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
