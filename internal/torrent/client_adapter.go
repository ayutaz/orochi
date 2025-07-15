package torrent

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/database"
	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
	torrentclient "github.com/ayutaz/orochi/internal/torrent_client"
)

// ClientAdapter adapts torrent_client.Client to the Manager interface.
type ClientAdapter struct {
	client  *torrentclient.Client
	logger  logger.Logger
	db      *database.DB
	updater *ProgressUpdater
}

// NewClientAdapter creates a new adapter for the torrent client.
func NewClientAdapter(cfg *config.Config, log logger.Logger) (*ClientAdapter, error) {
	client, err := torrentclient.NewClient(cfg, log)
	if err != nil {
		return nil, err
	}

	// Initialize database
	dbPath := cfg.DataDir + "/orochi.db"
	db, err := database.NewDB(dbPath, log)
	if err != nil {
		return nil, err
	}

	adapter := &ClientAdapter{
		client: client,
		logger: log,
		db:     db,
	}

	// Create and start progress updater
	adapter.updater = NewProgressUpdater(adapter, db, log)
	adapter.updater.Start()

	// Restore torrents from database
	if err := adapter.restoreTorrents(); err != nil {
		log.Error("failed to restore torrents", logger.Err(err))
	}

	return adapter, nil
}

// restoreTorrents restores torrents from the database.
func (a *ClientAdapter) restoreTorrents() error {
	records, err := a.db.ListTorrents()
	if err != nil {
		return err
	}

	for _, record := range records {
		// Decode metadata
		data, err := base64.StdEncoding.DecodeString(record.Metadata)
		if err != nil {
			a.logger.Error("failed to decode torrent metadata",
				logger.String("id", record.ID),
				logger.Err(err),
			)
			continue
		}

		// Add torrent back to client
		ctx := context.Background()
		_, err = a.client.AddTorrent(ctx, data)
		if err != nil {
			a.logger.Error("failed to restore torrent",
				logger.String("id", record.ID),
				logger.Err(err),
			)
			continue
		}

		a.logger.Info("restored torrent",
			logger.String("id", record.ID),
			logger.String("name", record.Name),
		)
	}

	return nil
}

// AddTorrent implements Manager.
func (a *ClientAdapter) AddTorrent(data []byte) (string, error) {
	ctx := context.Background()
	torr, err := a.client.AddTorrent(ctx, data)
	if err != nil {
		return "", err
	}

	// Save to database
	record := &database.TorrentRecord{
		ID:           torr.InfoHash(),
		InfoHash:     torr.InfoHash(),
		Name:         torr.Name(),
		Size:         torr.Length(),
		Status:       "stopped",
		Progress:     0,
		Downloaded:   0,
		Uploaded:     0,
		DownloadPath: torr.SavePath(),
		AddedAt:      time.Now(),
		Metadata:     base64.StdEncoding.EncodeToString(data),
	}

	if err := a.db.SaveTorrent(record); err != nil {
		a.logger.Error("failed to save torrent to database", logger.Err(err))
		// Don't fail the operation, just log the error
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

	// Get stats for upload data
	stats := torr.GetStats()

	// Convert to our Torrent struct
	return &Torrent{
		ID:           torr.InfoHash(),
		Info:         a.createTorrentInfo(torr),
		Status:       a.mapStatus(torr.Status()),
		Progress:     torr.Progress(),
		Downloaded:   torr.BytesCompleted(),
		Uploaded:     stats.BytesWrittenData,
		DownloadRate: torr.DownloadRate(),
		UploadRate:   torr.UploadRate(),
		AddedAt:      torr.AddedAt(),
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

	// Collect all info hashes
	infoHashes := make([]string, 0, len(torrents))
	for _, torr := range torrents {
		infoHashes = append(infoHashes, torr.InfoHash())
	}

	// Batch fetch all torrents
	torrentMap, err := a.client.GetTorrentsBatch(infoHashes)
	if err != nil {
		a.logger.Error("failed to batch fetch torrents", logger.Err(err))
		// Fall back to individual fetching
		return a.listTorrentsFallback(torrents)
	}

	// Build result
	result := make([]*Torrent, 0, len(torrentMap))
	for infoHash, torr := range torrentMap {
		t := a.convertTorrent(torr, infoHash)
		if t != nil {
			result = append(result, t)
		}
	}

	return result
}

// listTorrentsFallback is a fallback method using individual fetches.
func (a *ClientAdapter) listTorrentsFallback(torrents []*torrentclient.Torrent) []*Torrent {
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

// convertTorrent converts a torrent client torrent to a domain torrent.
func (a *ClientAdapter) convertTorrent(torr *torrentclient.Torrent, infoHash string) *Torrent {
	stats := torr.Stats()
	status := StatusStopped

	// Determine status based on torrent state
	switch {
	case stats.ActivePeers > 0 && torr.Progress() < 1:
		status = StatusDownloading
	case torr.Progress() == 1 && stats.ActivePeers > 0:
		status = StatusSeeding
	case torr.Progress() == 1:
		status = StatusStopped
	}

	// Get additional info from database
	dbRecord, _ := a.db.GetTorrent(infoHash)

	// Build TorrentInfo
	info := &TorrentInfo{
		InfoHash:    infoHash,
		Name:        torr.Name(),
		Length:      torr.Length(),
		PieceLength: int64(torr.PieceLength()),
		Files:       a.convertFiles(torr),
		Trackers:    a.getTrackers(dbRecord),
	}

	// Calculate download/upload stats
	downloaded := int64(0)
	uploaded := int64(0)
	if dbRecord != nil {
		downloaded = dbRecord.Downloaded
		uploaded = dbRecord.Uploaded
	}

	return &Torrent{
		ID:           infoHash,
		Info:         info,
		Status:       status,
		Progress:     torr.Progress(),
		Downloaded:   downloaded,
		Uploaded:     uploaded,
		DownloadRate: torr.DownloadRate(),
		UploadRate:   torr.UploadRate(),
		AddedAt:      time.Now(),
		Error:        "",
	}
}

// convertFiles converts torrent files to domain file info.
func (a *ClientAdapter) convertFiles(torr *torrentclient.Torrent) []FileInfo {
	files := torr.Files()
	fileInfos := make([]FileInfo, len(files))

	for i, f := range files {
		fileInfos[i] = FileInfo{
			Path:   strings.Split(f.Path, "/"),
			Length: f.Length,
		}
	}

	return fileInfos
}

// getTrackers extracts tracker info from database record.
func (a *ClientAdapter) getTrackers(dbRecord *database.TorrentRecord) []string {
	if dbRecord == nil {
		return []string{}
	}

	// TODO: Parse trackers from metadata or store separately
	return []string{}
}

// RemoveTorrent implements Manager.
func (a *ClientAdapter) RemoveTorrent(id string) error {
	torr, err := a.client.GetTorrent(id)
	if err != nil {
		return err
	}

	// Remove from database
	if err := a.db.DeleteTorrent(id); err != nil {
		a.logger.Error("failed to delete torrent from database", logger.Err(err))
		// Don't fail the operation, just log the error
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
	// Stop progress updater
	if a.updater != nil {
		a.updater.Stop()
	}

	// Close database
	if a.db != nil {
		_ = a.db.Close()
	}

	return a.client.Close()
}

// GetClient returns the underlying torrent client.
func (a *ClientAdapter) GetClient() (*torrentclient.Client, error) {
	if a.client == nil {
		return nil, errors.InternalErrorf("torrent client not initialized")
	}
	return a.client, nil
}

// GetDB returns the database connection.
func (a *ClientAdapter) GetDB() *database.DB {
	return a.db
}
