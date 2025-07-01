package torrent

import (
	"sync"
	"time"
	
	"github.com/ayutaz/orochi/internal/errors"
)

// TorrentStatus represents the status of a torrent.
type TorrentStatus string

const (
	// StatusStopped indicates the torrent is stopped.
	StatusStopped TorrentStatus = "stopped"
	// StatusDownloading indicates the torrent is downloading.
	StatusDownloading TorrentStatus = "downloading"
	// StatusSeeding indicates the torrent is seeding.
	StatusSeeding TorrentStatus = "seeding"
	// StatusError indicates the torrent has an error.
	StatusError TorrentStatus = "error"
)

// Torrent represents a managed torrent.
type Torrent struct {
	ID         string         `json:"id"`
	Info       *TorrentInfo   `json:"info"`
	Status     TorrentStatus  `json:"status"`
	Progress   float64        `json:"progress"`
	Downloaded int64          `json:"downloaded"`
	Uploaded   int64          `json:"uploaded"`
	AddedAt    time.Time      `json:"added_at"`
	Error      string         `json:"error,omitempty"`
}

// manager implements the Manager interface.
type manager struct {
	mu       sync.RWMutex
	torrents map[string]*Torrent
}

// NewManager creates a new torrent manager.
func NewManager() Manager {
	return &manager{
		torrents: make(map[string]*Torrent),
	}
}

// AddTorrent adds a torrent from torrent file data.
func (m *manager) AddTorrent(data []byte) (string, error) {
	info, err := ParseTorrentFile(data)
	if err != nil {
		return "", err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if torrent already exists
	if existing, exists := m.torrents[info.InfoHash]; exists {
		return existing.ID, nil
	}
	
	// Create new torrent
	torrent := &Torrent{
		ID:       info.InfoHash,
		Info:     info,
		Status:   StatusStopped,
		Progress: 0,
		AddedAt:  time.Now(),
	}
	
	m.torrents[info.InfoHash] = torrent
	
	return info.InfoHash, nil
}

// AddMagnet adds a torrent from a magnet link.
func (m *manager) AddMagnet(magnetLink string) (string, error) {
	info, err := ParseMagnetLink(magnetLink)
	if err != nil {
		return "", err
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if torrent already exists
	if existing, exists := m.torrents[info.InfoHash]; exists {
		return existing.ID, nil
	}
	
	// Create new torrent
	torrent := &Torrent{
		ID:       info.InfoHash,
		Info:     info,
		Status:   StatusStopped,
		Progress: 0,
		AddedAt:  time.Now(),
	}
	
	m.torrents[info.InfoHash] = torrent
	
	return info.InfoHash, nil
}

// RemoveTorrent removes a torrent.
func (m *manager) RemoveTorrent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.torrents[id]; !exists {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	// TODO: Stop torrent if running
	
	delete(m.torrents, id)
	
	return nil
}

// GetTorrent returns a torrent by ID.
func (m *manager) GetTorrent(id string) (*Torrent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	torrent, exists := m.torrents[id]
	return torrent, exists
}

// ListTorrents returns all torrents.
func (m *manager) ListTorrents() []*Torrent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	torrents := make([]*Torrent, 0, len(m.torrents))
	for _, torrent := range m.torrents {
		torrents = append(torrents, torrent)
	}
	
	return torrents
}

// Count returns the number of torrents.
func (m *manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.torrents)
}

// StartTorrent starts downloading a torrent.
func (m *manager) StartTorrent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	torrent, exists := m.torrents[id]
	if !exists {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	// TODO: Actually start the torrent download
	torrent.Status = StatusDownloading
	
	return nil
}

// StopTorrent stops a torrent.
func (m *manager) StopTorrent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	torrent, exists := m.torrents[id]
	if !exists {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	// TODO: Actually stop the torrent
	torrent.Status = StatusStopped
	
	return nil
}

// UpdateProgress updates the progress of a torrent.
func (t *Torrent) UpdateProgress(downloaded, uploaded int64) {
	t.Downloaded = downloaded
	t.Uploaded = uploaded
	
	if t.Info.Length > 0 {
		t.Progress = float64(downloaded) / float64(t.Info.Length) * 100
	}
}