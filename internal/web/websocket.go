package web

import (
	"net/http"
	"time"

	"github.com/ayutaz/orochi/internal/logger"
	"github.com/gorilla/websocket"
)

// createUpgrader creates a websocket upgrader with appropriate CORS settings.
func createUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// If no origins specified, allow all (development mode)
			if len(allowedOrigins) == 0 {
				return true
			}
			
			// Check if request origin is in allowed list
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			
			return false
		},
	}
}

// Message represents a message sent over WebSocket.
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub manages WebSocket connections.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	logger     logger.Logger
	upgrader   websocket.Upgrader
}

// NewHub creates a new WebSocket hub.
func NewHub(log logger.Logger, allowedOrigins []string) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan Message),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		logger:     log,
		upgrader:   createUpgrader(allowedOrigins),
	}
}

// Run starts the WebSocket hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.Info("WebSocket client connected", logger.Int("total_clients", len(h.clients)))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				h.logger.Info("WebSocket client disconnected", logger.Int("total_clients", len(h.clients)))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				if err := client.WriteJSON(message); err != nil {
					h.logger.Error("Failed to send WebSocket message", logger.Err(err))
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastTorrentUpdate sends a torrent update to all connected clients.
func (h *Hub) BroadcastTorrentUpdate() {
	h.broadcast <- Message{
		Type: "torrent_update",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}
}

// BroadcastTorrentData sends actual torrent data to all connected clients.
func (h *Hub) BroadcastTorrentData(torrents interface{}) {
	h.broadcast <- Message{
		Type: "torrents",
		Data: torrents,
	}
}

// handleWebSocket handles WebSocket connections.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsHub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade WebSocket", logger.Err(err))
		return
	}

	s.wsHub.register <- conn

	// Keep connection alive
	go func() {
		defer func() {
			s.wsHub.unregister <- conn
		}()

		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					s.logger.Error("WebSocket error", logger.Err(err))
				}
				break
			}
		}
	}()

	// Send ping messages
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()
}
