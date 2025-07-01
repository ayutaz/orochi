package web

import (
	"net/http"
	"time"

	"github.com/ayutaz/orochi/internal/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// TODO: Restrict in production
		return true
	},
}

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocketHub manages WebSocket connections.
type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	logger     logger.Logger
}

// NewWebSocketHub creates a new WebSocket hub.
func NewWebSocketHub(log logger.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketMessage),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		logger:     log,
	}
}

// Run starts the WebSocket hub.
func (h *WebSocketHub) Run() {
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
func (h *WebSocketHub) BroadcastTorrentUpdate() {
	h.broadcast <- WebSocketMessage{
		Type: "torrent_update",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}
}

// BroadcastTorrentData sends actual torrent data to all connected clients.
func (h *WebSocketHub) BroadcastTorrentData(torrents interface{}) {
	h.broadcast <- WebSocketMessage{
		Type: "torrents",
		Data: torrents,
	}
}

// handleWebSocket handles WebSocket connections.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
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

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}