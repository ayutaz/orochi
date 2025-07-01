package torrent

import (
	"context"
	"strings"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/logger"
	torrentclient "github.com/ayutaz/orochi/internal/torrent_client"
)

// ClientAdapter adapts torrent_client.Client to the Manager interface.
type ClientAdapter struct {
	client *torrentclient.Client
	logger logger.Logger
}

// NewClientAdapter creates a new adapter for the torrent client.
func NewClientAdapter(cfg *config.Config, log logger.Logger) (*ClientAdapter, error) {
	client, err := torrentclient.NewClient(cfg, log)
	if err != nil {
		return nil, err
	}

	return &ClientAdapter{
		client: client,
		logger: log,
	}, nil
}

// AddTorrent implements Manager.
func (a *ClientAdapter) AddTorrent(data []byte) (string, error) {
	ctx := context.Background()
	torr, err := a.client.AddTorrent(ctx, data)
	if err != nil {
		return "", err
	}
	return torr.InfoHash(), nil
}

// AddMagnet implements Manager.
func (a *ClientAdapter) AddMagnet(magnetLink string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	torr, err := a.client.AddMagnet(ctx, magnetLink)
	if err != nil {
		return "", err
	}
	return torr.InfoHash(), nil
}

// GetTorrent implements Manager.
func (a *ClientAdapter) GetTorrent(id string) (*Torrent, bool) {
	torr, err := a.client.GetTorrent(id)
	if err != nil {
		return nil, false
	}

	// Convert to our Torrent struct
	return &Torrent{
		ID:           torr.InfoHash(),
		Info:         a.createTorrentInfo(torr),
		Status:       a.mapStatus(torr.Status()),
		Progress:     torr.Progress(),
		Downloaded:   torr.BytesCompleted(),
		Uploaded:     0, // TODO: get from stats
		DownloadRate: torr.DownloadRate(),
		UploadRate:   torr.UploadRate(),
		AddedAt:      time.Now(), // TODO: track this
		Error:        "",
	}, true
}

// mapStatus converts client status to our Status type.
func (a *ClientAdapter) mapStatus(status string) Status {
	switch status {
	case "downloading":
		return StatusDownloading
	case "seeding":
		return StatusSeeding
	case "stopped":
		return StatusStopped
	default:
		return StatusStopped
	}
}

// createTorrentInfo creates a TorrentInfo from the torrent client torrent.
func (a *ClientAdapter) createTorrentInfo(torr *torrentclient.Torrent) *TorrentInfo {
	// Convert files
	files := torr.Files()
	fileInfos := make([]FileInfo, len(files))
	for i, f := range files {
		fileInfos[i] = FileInfo{
			Path:   strings.Split(f.Path, "/"),
			Length: f.Length,
		}
	}

	return &TorrentInfo{
		InfoHash:    torr.InfoHash(),
		Name:        torr.Name(),
		Length:      torr.Length(),
		PieceLength: 0,          // TODO: get from torrent
		Announce:    "",         // TODO: get from torrent
		Trackers:    []string{}, // TODO: get from torrent
		Files:       fileInfos,
	}
}

// ListTorrents implements Manager.
func (a *ClientAdapter) ListTorrents() []*Torrent {
	torrents := a.client.ListTorrents()
	result := make([]*Torrent, 0, len(torrents))

	for _, torr := range torrents {
		t, ok := a.GetTorrent(torr.InfoHash())
		if !ok {
			a.logger.Error("failed to get torrent details",
				logger.String("info_hash", torr.InfoHash()),
			)
			continue
		}
		result = append(result, t)
	}

	return result
}

// RemoveTorrent implements Manager.
func (a *ClientAdapter) RemoveTorrent(id string) error {
	torr, err := a.client.GetTorrent(id)
	if err != nil {
		return err
	}
	return torr.Remove()
}

// StartTorrent implements Manager.
func (a *ClientAdapter) StartTorrent(id string) error {
	torr, err := a.client.GetTorrent(id)
	if err != nil {
		return err
	}
	torr.Start()
	return nil
}

// StopTorrent implements Manager.
func (a *ClientAdapter) StopTorrent(id string) error {
	torr, err := a.client.GetTorrent(id)
	if err != nil {
		return err
	}
	torr.Stop()
	return nil
}

// Count implements Manager.
func (a *ClientAdapter) Count() int {
	return len(a.client.ListTorrents())
}

// Close closes the adapter and underlying client.
func (a *ClientAdapter) Close() error {
	return a.client.Close()
}
