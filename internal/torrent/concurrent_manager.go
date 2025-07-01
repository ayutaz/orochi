package torrent

import (
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/ayutaz/orochi/internal/errors"
)

// ConcurrentManager is a highly concurrent implementation of the Manager interface
type ConcurrentManager struct {
	torrents sync.Map   // map[string]*Torrent
	count    int64      // Atomic counter for torrent count
}

// NewConcurrentManager creates a new concurrent torrent manager
func NewConcurrentManager() Manager {
	return &ConcurrentManager{}
}

// AddTorrent adds a torrent from torrent file data
func (m *ConcurrentManager) AddTorrent(data []byte) (string, error) {
	info, err := ParseTorrentFile(data)
	if err != nil {
		return "", err
	}
	
	// Create new torrent
	torrent := &Torrent{
		ID:       info.InfoHash,
		Info:     info,
		Status:   StatusStopped,
		Progress: 0,
		AddedAt:  time.Now(),
	}
	
	// Check if already exists
	if _, loaded := m.torrents.LoadOrStore(info.InfoHash, torrent); loaded {
		// Already exists, return existing ID
		return info.InfoHash, nil
	}
	
	// Increment count
	atomic.AddInt64(&m.count, 1)
	
	return info.InfoHash, nil
}

// AddMagnet adds a torrent from a magnet link
func (m *ConcurrentManager) AddMagnet(magnetLink string) (string, error) {
	info, err := ParseMagnetLink(magnetLink)
	if err != nil {
		return "", err
	}
	
	// Create new torrent
	torrent := &Torrent{
		ID:       info.InfoHash,
		Info:     info,
		Status:   StatusStopped,
		Progress: 0,
		AddedAt:  time.Now(),
	}
	
	// Check if already exists
	if _, loaded := m.torrents.LoadOrStore(info.InfoHash, torrent); loaded {
		// Already exists, return existing ID
		return info.InfoHash, nil
	}
	
	// Increment count
	atomic.AddInt64(&m.count, 1)
	
	return info.InfoHash, nil
}

// RemoveTorrent removes a torrent
func (m *ConcurrentManager) RemoveTorrent(id string) error {
	if _, loaded := m.torrents.LoadAndDelete(id); !loaded {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	// Decrement count
	atomic.AddInt64(&m.count, -1)
	
	return nil
}

// GetTorrent returns a torrent by ID
func (m *ConcurrentManager) GetTorrent(id string) (*Torrent, bool) {
	value, exists := m.torrents.Load(id)
	if !exists {
		return nil, false
	}
	torrent, ok := value.(*Torrent)
	if !ok {
		return nil, false
	}
	return torrent, true
}

// ListTorrents returns all torrents
func (m *ConcurrentManager) ListTorrents() []*Torrent {
	// Pre-allocate slice with approximate size
	count := atomic.LoadInt64(&m.count)
	torrents := make([]*Torrent, 0, count)
	
	m.torrents.Range(func(_, value interface{}) bool {
		if torrent, ok := value.(*Torrent); ok {
			torrents = append(torrents, torrent)
		}
		return true
	})
	
	return torrents
}

// Count returns the number of torrents
func (m *ConcurrentManager) Count() int {
	return int(atomic.LoadInt64(&m.count))
}

// StartTorrent starts downloading a torrent
func (m *ConcurrentManager) StartTorrent(id string) error {
	value, exists := m.torrents.Load(id)
	if !exists {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	torrent, ok := value.(*Torrent)
	if !ok {
		return errors.InternalWithError("invalid torrent type", nil)
	}
	// In real implementation, this would need proper synchronization
	// For now, we're just updating the status
	torrent.Status = StatusDownloading
	
	return nil
}

// StopTorrent stops a torrent
func (m *ConcurrentManager) StopTorrent(id string) error {
	value, exists := m.torrents.Load(id)
	if !exists {
		return errors.NotFoundf("torrent with id %s not found", id)
	}
	
	torrent, ok := value.(*Torrent)
	if !ok {
		return errors.InternalWithError("invalid torrent type", nil)
	}
	// In real implementation, this would need proper synchronization
	// For now, we're just updating the status
	torrent.Status = StatusStopped
	
	return nil
}