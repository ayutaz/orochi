package web

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/ayutaz/orochi/internal/torrent"
)

// APIError represents an API error response.
type APIError struct {
	Error string `json:"error"`
}

// TorrentResponse represents a torrent in API responses.
type TorrentResponse struct {
	ID           string                 `json:"id"`
	Info         torrent.TorrentInfo    `json:"info"`
	Status       string                 `json:"status"`
	Progress     float64                `json:"progress"`
	Downloaded   int64                  `json:"downloaded"`
	Uploaded     int64                  `json:"uploaded"`
	DownloadRate int64                  `json:"downloadRate"`
	UploadRate   int64                  `json:"uploadRate"`
	AddedAt      string                 `json:"addedAt"`
	Error        string                 `json:"error,omitempty"`
}

// toTorrentResponse converts a torrent to API response format.
func toTorrentResponse(t *torrent.Torrent) TorrentResponse {
	return TorrentResponse{
		ID:           t.ID,
		Info:         *t.Info,
		Status:       string(t.Status),
		Progress:     t.Progress,
		Downloaded:   t.Downloaded,
		Uploaded:     t.Uploaded,
		DownloadRate: t.DownloadRate,
		UploadRate:   t.UploadRate,
		AddedAt:      t.AddedAt.Format(time.RFC3339),
		Error:        t.Error,
	}
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

// handleListTorrents handles GET /api/torrents.
func (s *Server) handleListTorrents(w http.ResponseWriter, _ *http.Request) {
	torrents := s.torrentManager.ListTorrents()
	
	// Convert to API response format
	responses := make([]TorrentResponse, 0, len(torrents))
	for _, t := range torrents {
		responses = append(responses, toTorrentResponse(t))
	}
	
	_ = writeJSON(w, http.StatusOK, responses)
}

// handleAddTorrent handles POST /api/torrents.
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

// handleAddMagnet handles POST /api/torrents/magnet.
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

// handleGetTorrent handles GET /api/torrents/:id.
func (s *Server) handleGetTorrent(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]

	torrentObj, exists := s.torrentManager.GetTorrent(id)
	if !exists {
		writeError(w, http.StatusNotFound, "torrent not found")
		return
	}
	
	// Convert to API response format
	response := toTorrentResponse(torrentObj)
	_ = writeJSON(w, http.StatusOK, response)
}

// handleDeleteTorrent handles DELETE /api/torrents/:id.
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

// handleStartTorrent handles POST /api/torrents/:id/start.
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
	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// handleStopTorrent handles POST /api/torrents/:id/stop.
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
	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// handleGetSettings handles GET /api/settings.
func (s *Server) handleGetSettings(w http.ResponseWriter, _ *http.Request) {
	settings := map[string]interface{}{
		"language":       "ja",
		"theme":          "light",
		"downloadPath":   s.config.DownloadDir,
		"maxConnections": 200,
		"port":           s.config.Port,
	}
	_ = writeJSON(w, http.StatusOK, settings)
}

// handleUpdateSettings handles PUT /api/settings.
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		s.logger.Error("failed to decode settings", logger.Err(err))
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Here you would update the actual settings
	// For now, just return success
	s.logger.Info("settings updated")
	_ = writeJSON(w, http.StatusOK, settings)
}

// FileUpdateRequest represents a file update request.
type FileUpdateRequest struct {
	Files []struct {
		Path     string `json:"path"`
		Selected bool   `json:"selected"`
		Priority string `json:"priority,omitempty"`
	} `json:"files"`
}

// handleUpdateFiles handles PUT /api/torrents/:id/files.
func (s *Server) handleUpdateFiles(w http.ResponseWriter, r *http.Request) {
	params := GetParams(r)
	id := params["id"]

	// Check if torrent exists
	_, exists := s.torrentManager.GetTorrent(id)
	if !exists {
		writeError(w, http.StatusNotFound, "torrent not found")
		return
	}

	var req FileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode request", logger.Err(err))
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// TODO: Implement actual file selection in torrent client
	// For now, just log and return success
	s.logger.Info("file selection updated",
		logger.String("torrent_id", id),
		logger.Int("file_count", len(req.Files)),
	)

	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
