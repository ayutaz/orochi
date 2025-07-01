package performance

import (
	"bytes"
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

// BenchmarkServer tests server performance
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
			resp, err := http.Get(ts.URL + "/health")
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()
		}
	})
	
	b.Run("ListTorrents", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resp, err := http.Get(ts.URL + "/api/torrents")
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
			data, _ := json.Marshal(reqBody)
			
			resp, err := http.Post(ts.URL+"/api/torrents/magnet", "application/json", bytes.NewReader(data))
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
				
				switch i % 3 {
				case 0:
					resp, err = http.Get(ts.URL + "/health")
				case 1:
					resp, err = http.Get(ts.URL + "/api/torrents")
				case 2:
					resp, err = http.Get(ts.URL + "/metrics")
				}
				
				if err != nil {
					b.Fatal(err)
				}
				resp.Body.Close()
				i++
			}
		})
	})
}

// BenchmarkTorrentManager tests torrent manager performance
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
		manager.AddMagnet(magnet)
	}
	
	b.Run("Add", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=bench%d.txt", b.N+i, i)
			manager.AddMagnet(magnet)
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
					manager.AddMagnet(magnet)
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

// TestLoadPerformance tests performance under high load
func TestLoadPerformance(t *testing.T) {
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
					resp, err := http.Get(ts.URL + "/health")
					if err == nil {
						resp.Body.Close()
					}
					
				case 1:
					// List torrents
					resp, err := http.Get(ts.URL + "/api/torrents")
					if err == nil {
						resp.Body.Close()
					}
					
				case 2:
					// Add magnet
					magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=load%d-%d.txt", clientID*1000+req, clientID, req)
					reqBody := map[string]string{"magnet": magnet}
					data, _ := json.Marshal(reqBody)
					
					resp, err := http.Post(ts.URL+"/api/torrents/magnet", "application/json", bytes.NewReader(data))
					if err == nil {
						resp.Body.Close()
					}
					
				case 3:
					// Get torrent (might 404)
					id := fmt.Sprintf("%040d", req%50)
					resp, err := http.Get(ts.URL + "/api/torrents/" + id)
					if err == nil {
						resp.Body.Close()
					}
					
				case 4:
					// Metrics
					resp, err := http.Get(ts.URL + "/metrics")
					if err == nil {
						resp.Body.Close()
					}
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

// TestMemoryUsage tests for memory leaks
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
			manager.RemoveTorrent(id)
		}
		
		// Check count is back to 0
		if count := manager.Count(); count != 0 {
			t.Errorf("cycle %d: expected count 0, got %d", cycle, count)
		}
	}
}

// BenchmarkFileUpload tests file upload performance
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
		
		if _, err := part.Write(torrentData); err != nil {
			b.Fatal(err)
		}
		
		if err := writer.Close(); err != nil {
			b.Fatal(err)
		}
		
		// Send request
		resp, err := http.Post(ts.URL+"/api/torrents", writer.FormDataContentType(), &buf)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}