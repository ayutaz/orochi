package web

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ayutaz/orochi/internal/logger"
)

func TestWebSocketHub(t *testing.T) {
	t.Run("NewHub", func(t *testing.T) {
		log := logger.NewTest()
		hub := NewHub(log, []string{"http://localhost:3000"})

		if hub == nil {
			t.Fatal("NewHub returned nil")
		}

		if hub.clients == nil {
			t.Error("clients map not initialized")
		}

		if hub.broadcast == nil {
			t.Error("broadcast channel not initialized")
		}

		if hub.register == nil {
			t.Error("register channel not initialized")
		}

		if hub.unregister == nil {
			t.Error("unregister channel not initialized")
		}
	})

	t.Run("CORS check", func(t *testing.T) {
		tests := []struct {
			name           string
			allowedOrigins []string
			requestOrigin  string
			shouldAllow    bool
		}{
			{
				name:           "No origins specified allows all",
				allowedOrigins: []string{},
				requestOrigin:  "http://example.com",
				shouldAllow:    true,
			},
			{
				name:           "Allowed origin",
				allowedOrigins: []string{"http://localhost:3000"},
				requestOrigin:  "http://localhost:3000",
				shouldAllow:    true,
			},
			{
				name:           "Disallowed origin",
				allowedOrigins: []string{"http://localhost:3000"},
				requestOrigin:  "http://evil.com",
				shouldAllow:    false,
			},
			{
				name:           "Multiple allowed origins",
				allowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
				requestOrigin:  "http://localhost:8080",
				shouldAllow:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				upgrader := createUpgrader(tt.allowedOrigins)
				req := httptest.NewRequest("GET", "/ws", nil)
				req.Header.Set("Origin", tt.requestOrigin)

				result := upgrader.CheckOrigin(req)
				if result != tt.shouldAllow {
					t.Errorf("CheckOrigin() = %v, want %v", result, tt.shouldAllow)
				}
			})
		}
	})

	t.Run("BroadcastTorrentUpdate", func(t *testing.T) {
		log := logger.NewTest()
		hub := NewHub(log, nil)

		// Test that BroadcastTorrentUpdate creates correct message
		go func() {
			msg := <-hub.broadcast
			if msg.Type != "torrent_update" {
				t.Errorf("unexpected message type: %s", msg.Type)
			}
			data, ok := msg.Data.(map[string]interface{})
			if !ok {
				t.Error("message data is not a map")
			}
			if _, hasTimestamp := data["timestamp"]; !hasTimestamp {
				t.Error("message data missing timestamp")
			}
		}()

		// Send broadcast
		hub.BroadcastTorrentUpdate()

		// Give time for goroutine to process
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("BroadcastTorrentData", func(t *testing.T) {
		log := logger.NewTest()
		hub := NewHub(log, nil)

		testData := []map[string]interface{}{
			{"id": "123", "name": "test.torrent"},
		}

		// Test that BroadcastTorrentData creates correct message
		go func() {
			msg := <-hub.broadcast
			if msg.Type != "torrents" {
				t.Errorf("unexpected message type: %s", msg.Type)
			}
			if msg.Data == nil {
				t.Error("message data is nil")
			}
		}()

		// Send broadcast
		hub.BroadcastTorrentData(testData)

		// Give time for goroutine to process
		time.Sleep(50 * time.Millisecond)
	})
}

func TestWebSocketConnection(t *testing.T) {
	t.Skip("Skipping WebSocket connection tests - httptest doesn't support hijacking")
}

func TestWebSocketError(t *testing.T) {
	t.Skip("Skipping WebSocket error tests - httptest doesn't support hijacking")
}
