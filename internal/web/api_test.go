package web

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/torrent"
)

func TestAPI_Torrents(t *testing.T) {
	cfg := &config.Config{Port: 8080}
	server := NewServer(cfg)
	manager := torrent.NewManager()
	server.SetTorrentManager(manager)

	t.Run("GET /api/torrents - 空のリストを返す", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/torrents", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var torrents []torrent.Torrent
		if err := json.NewDecoder(w.Body).Decode(&torrents); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(torrents) != 0 {
			t.Errorf("expected 0 torrents, got %d", len(torrents))
		}
	})

	t.Run("POST /api/torrents - torrentファイルを追加", func(t *testing.T) {
		// Create multipart form data with torrent file
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		
		part, err := writer.CreateFormFile("torrent", "test.torrent")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		
		torrentData := torrent.CreateTestTorrent()
		if _, err := part.Write(torrentData); err != nil {
			t.Fatalf("failed to write torrent data: %v", err)
		}
		
		if err := writer.Close(); err != nil {
			t.Fatalf("failed to close writer: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/torrents", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response["id"] == "" {
			t.Error("expected non-empty ID in response")
		}
	})

	t.Run("POST /api/torrents/magnet - マグネットリンクを追加", func(t *testing.T) {
		body := map[string]string{
			"magnet": "magnet:?xt=urn:btih:1234567890abcdef1234567890abcdef12345678&dn=test.txt",
		}
		
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/torrents/magnet", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response["id"] != "1234567890abcdef1234567890abcdef12345678" {
			t.Errorf("expected ID to match info hash, got %v", response["id"])
		}
	})
}

func TestAPI_TorrentOperations(t *testing.T) {
	cfg := &config.Config{Port: 8080}
	server := NewServer(cfg)
	manager := torrent.NewManager()
	server.SetTorrentManager(manager)

	// Add a test torrent
	torrentData := torrent.CreateTestTorrent()
	id, _ := manager.AddTorrent(torrentData)

	t.Run("GET /api/torrents/:id - 特定のトレントを取得", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/torrents/"+id, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var torrentResp torrent.Torrent
		if err := json.NewDecoder(w.Body).Decode(&torrentResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if torrentResp.ID != id {
			t.Errorf("expected ID %s, got %s", id, torrentResp.ID)
		}
	})

	t.Run("POST /api/torrents/:id/start - トレントを開始", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/torrents/"+id+"/start", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify torrent status changed
		torrentObj, _ := manager.GetTorrent(id)
		if torrentObj.Status != torrent.StatusDownloading {
			t.Errorf("expected status %s, got %s", torrent.StatusDownloading, torrentObj.Status)
		}
	})

	t.Run("POST /api/torrents/:id/stop - トレントを停止", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/torrents/"+id+"/stop", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify torrent status changed
		torrentObj, _ := manager.GetTorrent(id)
		if torrentObj.Status != torrent.StatusStopped {
			t.Errorf("expected status %s, got %s", torrent.StatusStopped, torrentObj.Status)
		}
	})

	t.Run("DELETE /api/torrents/:id - トレントを削除", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/torrents/"+id, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		// Verify torrent was removed
		if _, exists := manager.GetTorrent(id); exists {
			t.Error("torrent should not exist after deletion")
		}
	})

	t.Run("GET /api/torrents/:id - 存在しないトレント", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/torrents/nonexistent", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}