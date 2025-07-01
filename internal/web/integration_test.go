package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	
	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/torrent"
)

func TestServerIntegration(t *testing.T) {
	// Create server
	cfg := config.LoadDefault()
	server := NewServer(cfg)
	
	// Set torrent manager
	manager := torrent.NewManager()
	server.SetTorrentManager(manager)
	
	// Create test server
	ts := httptest.NewServer(server.router)
	defer ts.Close()
	
	t.Run("Health check", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/health")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("expected body 'OK', got %s", body)
		}
	})
	
	t.Run("List torrents (empty)", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/torrents")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		var torrents []torrent.Torrent
		if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
			t.Fatal(err)
		}
		
		if len(torrents) != 0 {
			t.Errorf("expected 0 torrents, got %d", len(torrents))
		}
	})
	
	t.Run("Add torrent file", func(t *testing.T) {
		// Create multipart form
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		
		part, err := writer.CreateFormFile("torrent", "test.torrent")
		if err != nil {
			t.Fatal(err)
		}
		
		torrentData := torrent.CreateTestTorrent()
		if _, err := part.Write(torrentData); err != nil {
			t.Fatal(err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		
		// Send request
		resp, err := http.Post(ts.URL+"/api/torrents", writer.FormDataContentType(), &buf)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected status 201, got %d: %s", resp.StatusCode, body)
		}
		
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		
		if result["id"] == "" {
			t.Error("expected torrent ID in response")
		}
	})
	
	t.Run("Add magnet link", func(t *testing.T) {
		magnet := "magnet:?xt=urn:btih:1234567890abcdef1234567890abcdef12345678&dn=test.txt"
		reqBody := map[string]string{"magnet": magnet}
		
		data, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := http.Post(ts.URL+"/api/torrents/magnet", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected status 201, got %d: %s", resp.StatusCode, body)
		}
		
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		
		if result["id"] != "1234567890abcdef1234567890abcdef12345678" {
			t.Errorf("expected ID to match info hash, got %s", result["id"])
		}
	})
	
	t.Run("Get torrent", func(t *testing.T) {
		// First add a torrent
		magnet := "magnet:?xt=urn:btih:abcdef1234567890abcdef1234567890abcdef12&dn=get-test.txt"
		manager.AddMagnet(magnet)
		
		resp, err := http.Get(ts.URL + "/api/torrents/abcdef1234567890abcdef1234567890abcdef12")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		var torrentObj torrent.Torrent
		if err := json.NewDecoder(resp.Body).Decode(&torrentObj); err != nil {
			t.Fatal(err)
		}
		
		if torrentObj.ID != "abcdef1234567890abcdef1234567890abcdef12" {
			t.Errorf("expected torrent ID to match, got %s", torrentObj.ID)
		}
	})
	
	t.Run("Start torrent", func(t *testing.T) {
		// First add a torrent
		magnet := "magnet:?xt=urn:btih:fedcba0987654321fedcba0987654321fedcba09&dn=start-test.txt"
		id, _ := manager.AddMagnet(magnet)
		
		resp, err := http.Post(ts.URL+"/api/torrents/"+id+"/start", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		// Verify status changed
		torrentObj, _ := manager.GetTorrent(id)
		if torrentObj.Status != torrent.StatusDownloading {
			t.Errorf("expected status downloading, got %s", torrentObj.Status)
		}
	})
	
	t.Run("Stop torrent", func(t *testing.T) {
		// First add and start a torrent
		magnet := "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567&dn=stop-test.txt"
		id, _ := manager.AddMagnet(magnet)
		manager.StartTorrent(id)
		
		resp, err := http.Post(ts.URL+"/api/torrents/"+id+"/stop", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		
		// Verify status changed
		torrentObj, _ := manager.GetTorrent(id)
		if torrentObj.Status != torrent.StatusStopped {
			t.Errorf("expected status stopped, got %s", torrentObj.Status)
		}
	})
	
	t.Run("Delete torrent", func(t *testing.T) {
		// First add a torrent
		magnet := "magnet:?xt=urn:btih:deadbeef1234567890abcdefdeadbeef12345678&dn=delete-test.txt"
		id, _ := manager.AddMagnet(magnet)
		
		req, err := http.NewRequest("DELETE", ts.URL+"/api/torrents/"+id, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}
		
		// Verify deleted
		if _, exists := manager.GetTorrent(id); exists {
			t.Error("torrent should not exist after deletion")
		}
	})
	
	t.Run("404 for non-existent torrent", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/torrents/nonexistent")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
	
	t.Run("Invalid JSON", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/api/torrents/magnet", "application/json", strings.NewReader("invalid json"))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
	
	t.Run("Request ID header", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/health")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		reqID := resp.Header.Get("X-Request-ID")
		if reqID == "" {
			t.Error("expected X-Request-ID header")
		}
	})
}

func TestConcurrentRequests(t *testing.T) {
	// Create server
	cfg := config.LoadDefault()
	server := NewServer(cfg)
	server.SetTorrentManager(torrent.NewManager())
	
	// Create test server
	ts := httptest.NewServer(server.router)
	defer ts.Close()
	
	// Number of concurrent requests
	numRequests := 100
	done := make(chan bool, numRequests)
	
	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer func() { done <- true }()
			
			// Mix of different operations
			switch i % 4 {
			case 0:
				// List torrents
				resp, err := http.Get(ts.URL + "/api/torrents")
				if err != nil {
					t.Error(err)
					return
				}
				resp.Body.Close()
				
			case 1:
				// Add magnet
				magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=test%d.txt", i, i)
				reqBody := map[string]string{"magnet": magnet}
				data, _ := json.Marshal(reqBody)
				
				resp, err := http.Post(ts.URL+"/api/torrents/magnet", "application/json", bytes.NewReader(data))
				if err != nil {
					t.Error(err)
					return
				}
				resp.Body.Close()
				
			case 2:
				// Health check
				resp, err := http.Get(ts.URL + "/health")
				if err != nil {
					t.Error(err)
					return
				}
				resp.Body.Close()
				
			case 3:
				// 404 request
				resp, err := http.Get(ts.URL + "/api/torrents/nonexistent")
				if err != nil {
					t.Error(err)
					return
				}
				resp.Body.Close()
			}
		}(i)
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}