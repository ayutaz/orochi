package web

import (
	"encoding/json"
	"io"
	"net/http"
	
	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
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
	_ = writeJSON(w, status, APIError{Error: message})
}

// handleListTorrents handles GET /api/torrents
func (s *Server) handleListTorrents(w http.ResponseWriter, r *http.Request) {
	torrents := s.torrentManager.ListTorrents()
	_ = writeJSON(w, http.StatusOK, torrents)
}

// handleAddTorrent handles POST /api/torrents
func (s *Server) handleAddTorrent(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		s.logger.Error("failed to parse form data", logger.Err(err))
		writeError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, _, err := r.FormFile("torrent")
	if err != nil {
		s.logger.Error("torrent file missing", logger.Err(err))
		writeError(w, http.StatusBadRequest, "torrent file required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		s.logger.Error("failed to read torrent file", logger.Err(err))
		writeError(w, http.StatusBadRequest, "failed to read torrent file")
		return
	}

	id, err := s.torrentManager.AddTorrent(data)
	if err != nil {
		if errors.IsInvalidInput(err) || errors.IsParseError(err) {
			s.logger.Warn("invalid torrent file", logger.Err(err))
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			s.logger.Error("failed to add torrent", logger.Err(err))
			writeError(w, http.StatusInternalServerError, "failed to add torrent")
		}
		return
	}

	s.logger.Info("torrent added", logger.String("id", id))
	_ = writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// handleAddMagnet handles POST /api/torrents/magnet
func (s *Server) handleAddMagnet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Magnet string `json:"magnet"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode JSON", logger.Err(err))
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Magnet == "" {
		writeError(w, http.StatusBadRequest, "magnet link required")
		return
	}

	id, err := s.torrentManager.AddMagnet(req.Magnet)
	if err != nil {
		if errors.IsInvalidInput(err) || errors.IsParseError(err) {
			s.logger.Warn("invalid magnet link", logger.Err(err))
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			s.logger.Error("failed to add magnet", logger.Err(err))
			writeError(w, http.StatusInternalServerError, "failed to add magnet")
		}
		return
	}

	s.logger.Info("magnet added", logger.String("id", id))
	_ = writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// handleGetTorrent handles GET /api/torrents/:id
func (s *Server) handleGetTorrent(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]
	
	torrentObj, exists := s.torrentManager.GetTorrent(id)
	if !exists {
		writeError(w, http.StatusNotFound, "torrent not found")
		return
	}
	writeJSON(w, http.StatusOK, torrentObj)
}

// handleDeleteTorrent handles DELETE /api/torrents/:id
func (s *Server) handleDeleteTorrent(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]
	
	if err := s.torrentManager.RemoveTorrent(id); err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("torrent not found", logger.String("id", id))
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			s.logger.Error("failed to remove torrent", logger.String("id", id), logger.Err(err))
			writeError(w, http.StatusInternalServerError, "failed to remove torrent")
		}
		return
	}
	
	s.logger.Info("torrent removed", logger.String("id", id))
	w.WriteHeader(http.StatusNoContent)
}

// handleStartTorrent handles POST /api/torrents/:id/start
func (s *Server) handleStartTorrent(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]
	
	if err := s.torrentManager.StartTorrent(id); err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("torrent not found", logger.String("id", id))
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			s.logger.Error("failed to start torrent", logger.String("id", id), logger.Err(err))
			writeError(w, http.StatusInternalServerError, "failed to start torrent")
		}
		return
	}
	
	s.logger.Info("torrent started", logger.String("id", id))
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// handleStopTorrent handles POST /api/torrents/:id/stop
func (s *Server) handleStopTorrent(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]
	
	if err := s.torrentManager.StopTorrent(id); err != nil {
		if errors.IsNotFound(err) {
			s.logger.Warn("torrent not found", logger.String("id", id))
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			s.logger.Error("failed to stop torrent", logger.String("id", id), logger.Err(err))
			writeError(w, http.StatusInternalServerError, "failed to stop torrent")
		}
		return
	}
	
	s.logger.Info("torrent stopped", logger.String("id", id))
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}