package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	
	"github.com/ayutaz/orochi/internal/errors"
)

// APIError represents an API error response.
type APIError struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIError{Error: message})
}

// handleAPITorrents handles /api/torrents endpoint.
func (s *Server) handleAPITorrents(w http.ResponseWriter, r *http.Request) {
	if s.torrentManager == nil {
		writeError(w, http.StatusInternalServerError, "torrent manager not initialized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// List all torrents
		torrents := s.torrentManager.ListTorrents()
		writeJSON(w, http.StatusOK, torrents)

	case http.MethodPost:
		// Add a new torrent from file upload
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
			writeError(w, http.StatusBadRequest, "failed to parse form data")
			return
		}

		file, _, err := r.FormFile("torrent")
		if err != nil {
			writeError(w, http.StatusBadRequest, "torrent file required")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read torrent file")
			return
		}

		id, err := s.torrentManager.AddTorrent(data)
		if err != nil {
			if errors.IsInvalidInput(err) {
				writeError(w, http.StatusBadRequest, err.Error())
			} else {
				writeError(w, http.StatusInternalServerError, "failed to add torrent")
			}
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{"id": id})

	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAPITorrentMagnet handles /api/torrents/magnet endpoint.
func (s *Server) handleAPITorrentMagnet(w http.ResponseWriter, r *http.Request) {
	if s.torrentManager == nil {
		writeError(w, http.StatusInternalServerError, "torrent manager not initialized")
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Magnet string `json:"magnet"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Magnet == "" {
		writeError(w, http.StatusBadRequest, "magnet link required")
		return
	}

	id, err := s.torrentManager.AddMagnet(req.Magnet)
	if err != nil {
		if errors.IsInvalidInput(err) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to add magnet")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// handleAPITorrent handles /api/torrents/:id endpoint.
func (s *Server) handleAPITorrent(w http.ResponseWriter, r *http.Request) {
	if s.torrentManager == nil {
		writeError(w, http.StatusInternalServerError, "torrent manager not initialized")
		return
	}

	// Extract torrent ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/torrents/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "torrent ID required")
		return
	}

	id := parts[0]

	// Handle operations on specific torrent
	if len(parts) > 1 {
		operation := parts[1]
		s.handleTorrentOperation(w, r, id, operation)
		return
	}

	// Handle torrent CRUD
	switch r.Method {
	case http.MethodGet:
		// Get torrent details
		torrentObj, exists := s.torrentManager.GetTorrent(id)
		if !exists {
			writeError(w, http.StatusNotFound, "torrent not found")
			return
		}
		writeJSON(w, http.StatusOK, torrentObj)

	case http.MethodDelete:
		// Remove torrent
		if err := s.torrentManager.RemoveTorrent(id); err != nil {
			if errors.IsNotFound(err) {
				writeError(w, http.StatusNotFound, err.Error())
			} else {
				writeError(w, http.StatusInternalServerError, "failed to remove torrent")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleTorrentOperation handles operations on a specific torrent.
func (s *Server) handleTorrentOperation(w http.ResponseWriter, r *http.Request, id, operation string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	switch operation {
	case "start":
		if err := s.torrentManager.StartTorrent(id); err != nil {
			if errors.IsNotFound(err) {
				writeError(w, http.StatusNotFound, err.Error())
			} else {
				writeError(w, http.StatusInternalServerError, "failed to start torrent")
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "started"})

	case "stop":
		if err := s.torrentManager.StopTorrent(id); err != nil {
			if errors.IsNotFound(err) {
				writeError(w, http.StatusNotFound, err.Error())
			} else {
				writeError(w, http.StatusInternalServerError, "failed to stop torrent")
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})

	default:
		writeError(w, http.StatusBadRequest, "unknown operation")
	}
}