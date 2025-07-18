package torrentclient

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/ayutaz/orochi/internal/config"
	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
	"github.com/ayutaz/orochi/internal/network"
)

// Client wraps the anacrolix torrent client.
type Client struct {
	client         *torrent.Client
	logger         logger.Logger
	config         *config.Config
	networkMonitor *network.Monitor
}

// NewClient creates a new torrent client.
func NewClient(cfg *config.Config, log logger.Logger) (*Client, error) {
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = cfg.GetAbsoluteDownloadDir()
	// Use configured port for BitTorrent, or 0 for random (useful in tests)
	if cfg.Port == 0 {
		clientConfig.ListenPort = 0 // Random port
	} else {
		clientConfig.ListenPort = 6881 // BitTorrent default port
	}
	clientConfig.Seed = true
	// clientConfig.Logger = log // TODO: implement logger adapter

	// Create storage
	storageImpl := storage.NewFileByInfoHash(cfg.GetAbsoluteDownloadDir())
	clientConfig.DefaultStorage = storageImpl

	// Create torrent client
	torrentClient, err := torrent.NewClient(clientConfig)
	if err != nil {
		return nil, errors.InternalWithError("failed to create torrent client", err)
	}

	client := &Client{
		client: torrentClient,
		logger: log,
		config: cfg,
	}

	// Set up VPN monitoring if enabled
	if cfg.VPN != nil && cfg.VPN.Enabled {
		client.networkMonitor = network.NewNetworkMonitor(cfg.VPN)
		client.networkMonitor.SetLogger(log)
		client.networkMonitor.Start()

		log.Info("VPN binding enabled",
			logger.String("interface", cfg.VPN.InterfaceName),
			logger.Bool("kill_switch", cfg.VPN.KillSwitch),
		)
	}

	return client, nil
}

// AddTorrent adds a torrent from file data.
func (c *Client) AddTorrent(_ context.Context, data []byte) (*Torrent, error) {
	// Check VPN status if kill switch is enabled
	if c.networkMonitor != nil && !c.networkMonitor.ShouldAllowConnection() {
		return nil, errors.PermissionDeniedf("VPN kill switch active - VPN connection required")
	}

	// Parse torrent data
	metaInfo, err := metainfo.Load(bytes.NewReader(data))
	if err != nil {
		return nil, errors.ParseError("failed to parse torrent file", err)
	}

	// Add torrent to client
	t, err := c.client.AddTorrent(metaInfo)
	if err != nil {
		return nil, errors.InternalWithError("failed to add torrent", err)
	}

	// Start downloading
	t.DownloadAll()

	c.logger.Info("torrent added",
		logger.String("name", t.Name()),
		logger.String("info_hash", t.InfoHash().HexString()),
	)

	return &Torrent{
		torrent:    t,
		client:     c,
		addedAt:    time.Now(),
		lastUpdate: time.Now(),
	}, nil
}

// AddMagnet adds a torrent from a magnet link.
func (c *Client) AddMagnet(ctx context.Context, magnetLink string) (*Torrent, error) {
	// Check VPN status if kill switch is enabled
	if c.networkMonitor != nil && !c.networkMonitor.ShouldAllowConnection() {
		return nil, errors.PermissionDeniedf("VPN kill switch active - VPN connection required")
	}

	// Add magnet link
	t, err := c.client.AddMagnet(magnetLink)
	if err != nil {
		return nil, errors.ParseError("failed to add magnet link", err)
	}

	// Wait for info to be available
	select {
	case <-t.GotInfo():
		c.logger.Info("torrent metadata received",
			logger.String("name", t.Name()),
			logger.String("info_hash", t.InfoHash().HexString()),
		)
	case <-ctx.Done():
		t.Drop()
		return nil, errors.Timeout("timeout waiting for torrent metadata")
	}

	// Start downloading
	t.DownloadAll()

	return &Torrent{
		torrent:    t,
		client:     c,
		addedAt:    time.Now(),
		lastUpdate: time.Now(),
	}, nil
}

// GetTorrent returns a torrent by info hash.
func (c *Client) GetTorrent(infoHash string) (*Torrent, error) {
	// Validate info hash length
	if len(infoHash) != 40 {
		return nil, errors.InvalidInputf("invalid info hash length: %d (expected 40)", len(infoHash))
	}

	// Parse info hash
	ih := metainfo.NewHashFromHex(infoHash)

	// Find torrent
	t, ok := c.client.Torrent(ih)
	if !ok {
		return nil, errors.NotFoundf("torrent not found: %s", infoHash)
	}

	return &Torrent{
		torrent:    t,
		client:     c,
		addedAt:    time.Now(),
		lastUpdate: time.Now(),
	}, nil
}

// GetTorrentsBatch returns multiple torrents by their info hashes.
// This is more efficient than calling GetTorrent multiple times.
func (c *Client) GetTorrentsBatch(infoHashes []string) (map[string]*Torrent, error) {
	result := make(map[string]*Torrent, len(infoHashes))

	// Get all torrents once
	allTorrents := c.client.Torrents()
	torrentMap := make(map[metainfo.Hash]*torrent.Torrent, len(allTorrents))

	// Build a map for O(1) lookup
	for _, t := range allTorrents {
		torrentMap[t.InfoHash()] = t
	}

	// Look up requested torrents
	for _, infoHash := range infoHashes {
		// Validate info hash length
		if len(infoHash) != 40 {
			c.logger.Warn("invalid info hash length in batch",
				logger.String("info_hash", infoHash),
				logger.Int("length", len(infoHash)))
			continue
		}

		// Parse info hash
		ih := metainfo.NewHashFromHex(infoHash)

		// Find torrent
		if t, ok := torrentMap[ih]; ok {
			result[infoHash] = &Torrent{
				torrent:    t,
				client:     c,
				addedAt:    time.Now(),
				lastUpdate: time.Now(),
			}
		}
	}

	return result, nil
}

// ListTorrents returns all torrents.
func (c *Client) ListTorrents() []*Torrent {
	torrents := c.client.Torrents()
	result := make([]*Torrent, 0, len(torrents))

	for _, t := range torrents {
		result = append(result, &Torrent{
			torrent:    t,
			client:     c,
			addedAt:    time.Now(),
			lastUpdate: time.Now(),
		})
	}

	return result
}

// Close closes the torrent client.
func (c *Client) Close() error {
	// Stop network monitor if running
	if c.networkMonitor != nil {
		c.networkMonitor.Stop()
	}

	c.client.Close()
	return nil
}

// GetNetworkMonitor returns the network monitor if available.
func (c *Client) GetNetworkMonitor() *network.Monitor {
	return c.networkMonitor
}

// Torrent represents a torrent in the client.
type Torrent struct {
	torrent    *torrent.Torrent
	client     *Client
	addedAt    time.Time
	lastStats  torrent.TorrentStats
	lastUpdate time.Time
}

// InfoHash returns the torrent's info hash.
func (t *Torrent) InfoHash() string {
	return t.torrent.InfoHash().HexString()
}

// Name returns the torrent's name.
func (t *Torrent) Name() string {
	return t.torrent.Name()
}

// Length returns the torrent's total length in bytes.
func (t *Torrent) Length() int64 {
	return t.torrent.Length()
}

// PieceLength returns the length of each piece.
func (t *Torrent) PieceLength() int {
	info := t.torrent.Info()
	if info != nil {
		return int(info.PieceLength)
	}
	return 0
}

// NumPieces returns the number of pieces in the torrent.
func (t *Torrent) NumPieces() int {
	return t.torrent.NumPieces()
}

// CompletedPieces returns the number of completed pieces.
func (t *Torrent) CompletedPieces() int {
	stats := t.torrent.Stats()
	return stats.PiecesComplete
}

// BytesCompleted returns the number of bytes completed.
func (t *Torrent) BytesCompleted() int64 {
	return t.torrent.BytesCompleted()
}

// KnownPeers returns all known peers.
func (t *Torrent) KnownPeers() []torrent.PeerInfo {
	return t.torrent.KnownSwarm()
}

// Private returns whether the torrent is private.
func (t *Torrent) Private() bool {
	info := t.torrent.Info()
	if info != nil && info.Private != nil {
		return *info.Private
	}
	return false
}

// AddedAt returns when the torrent was added.
func (t *Torrent) AddedAt() time.Time {
	return t.addedAt
}

// CompletedAt returns when the torrent was completed.
func (t *Torrent) CompletedAt() *time.Time {
	if t.Progress() == 1 {
		completed := time.Now()
		return &completed
	}
	return nil
}

// Progress returns the download progress as a percentage (0-100).
func (t *Torrent) Progress() float64 {
	if t.torrent.Length() == 0 {
		return 0
	}
	return float64(t.torrent.BytesCompleted()) / float64(t.torrent.Length()) * 100
}

// Status returns the torrent's status.
func (t *Torrent) Status() string {
	if t.torrent.Seeding() {
		return "seeding"
	}
	if t.torrent.BytesCompleted() < t.torrent.Length() {
		return "downloading"
	}
	return "stopped"
}

// Stats returns the torrent's statistics.
func (t *Torrent) Stats() torrent.TorrentStats {
	return t.torrent.Stats()
}

// DownloadRate returns the current download rate in bytes per second.
func (t *Torrent) DownloadRate() int64 {
	currentStats := t.torrent.Stats()
	now := time.Now()

	// Calculate rate based on change since last update
	if t.lastUpdate.IsZero() || now.Sub(t.lastUpdate) < time.Second {
		// Not enough time has passed for accurate measurement
		return 0
	}

	timeDelta := now.Sub(t.lastUpdate).Seconds()
	bytesDelta := currentStats.BytesReadUsefulData.Int64() - t.lastStats.BytesReadUsefulData.Int64()

	// Update cached stats
	t.lastStats = currentStats
	t.lastUpdate = now

	if timeDelta > 0 {
		return int64(float64(bytesDelta) / timeDelta)
	}
	return 0
}

// UploadRate returns the current upload rate in bytes per second.
func (t *Torrent) UploadRate() int64 {
	currentStats := t.torrent.Stats()
	now := time.Now()

	// Calculate rate based on change since last update
	if t.lastUpdate.IsZero() || now.Sub(t.lastUpdate) < time.Second {
		// Not enough time has passed for accurate measurement
		return 0
	}

	timeDelta := now.Sub(t.lastUpdate).Seconds()
	bytesDelta := currentStats.BytesWrittenData.Int64() - t.lastStats.BytesWrittenData.Int64()

	if timeDelta > 0 {
		return int64(float64(bytesDelta) / timeDelta)
	}
	return 0
}

// Start starts downloading the torrent.
func (t *Torrent) Start() {
	// Check VPN status if kill switch is enabled
	if t.client.networkMonitor != nil && !t.client.networkMonitor.ShouldAllowConnection() {
		t.client.logger.Warn("torrent start blocked by VPN kill switch",
			logger.String("name", t.Name()),
			logger.String("info_hash", t.InfoHash()),
		)
		return
	}

	t.torrent.DownloadAll()
	t.client.logger.Info("torrent started",
		logger.String("name", t.Name()),
		logger.String("info_hash", t.InfoHash()),
	)
}

// Stop stops the torrent.
func (t *Torrent) Stop() {
	t.torrent.DisallowDataDownload()
	t.torrent.DisallowDataUpload()
	t.client.logger.Info("torrent stopped",
		logger.String("name", t.Name()),
		logger.String("info_hash", t.InfoHash()),
	)
}

// Remove removes the torrent.
func (t *Torrent) Remove() error {
	t.torrent.Drop()
	t.client.logger.Info("torrent removed",
		logger.String("name", t.Name()),
		logger.String("info_hash", t.InfoHash()),
	)
	return nil
}

// Files returns the torrent's files.
func (t *Torrent) Files() []File {
	files := t.torrent.Files()
	result := make([]File, 0, len(files))

	for _, f := range files {
		result = append(result, File{
			Path:   f.Path(),
			Length: f.Length(),
		})
	}

	return result
}

// SavePath returns the path where the torrent is saved.
func (t *Torrent) SavePath() string {
	return filepath.Join(t.client.config.GetAbsoluteDownloadDir(), t.torrent.Name())
}

// SetFilePriority sets the priority for a specific file.
func (t *Torrent) SetFilePriority(fileIndex, priority int) error {
	files := t.torrent.Files()
	if fileIndex < 0 || fileIndex >= len(files) {
		return errors.InvalidInputf("file index %d out of range", fileIndex)
	}

	// Set priority: 0 = don't download, 1 = normal, 2 = high
	// Convert int to PiecePriority type
	var piecePrio torrent.PiecePriority
	switch priority {
	case 0:
		piecePrio = torrent.PiecePriorityNone
	case 1:
		piecePrio = torrent.PiecePriorityNormal
	case 2:
		piecePrio = torrent.PiecePriorityHigh
	default:
		piecePrio = torrent.PiecePriorityNormal
	}
	files[fileIndex].SetPriority(piecePrio)
	return nil
}

// SetFileSelected sets whether a file should be downloaded.
func (t *Torrent) SetFileSelected(fileIndex int, selected bool) error {
	priority := 0
	if selected {
		priority = 1
	}
	return t.SetFilePriority(fileIndex, priority)
}

// GetStats returns torrent statistics in our format.
func (t *Torrent) GetStats() Stats {
	stats := t.torrent.Stats()
	return Stats{
		BytesRead:         stats.BytesRead.Int64(),
		BytesWritten:      stats.BytesWritten.Int64(),
		BytesReadData:     stats.BytesReadData.Int64(),
		BytesWrittenData:  stats.BytesWrittenData.Int64(),
		ChunksReadWasted:  stats.ChunksReadWasted.Int64(),
		ChunksWritten:     stats.ChunksWritten.Int64(),
		PiecesDirtiedGood: stats.PiecesDirtiedGood.Int64(),
		PiecesDirtiedBad:  stats.PiecesDirtiedBad.Int64(),
		ActivePeers:       len(t.torrent.PeerConns()),
		ConnectedSeeders:  t.countSeeders(),
		TotalPeers:        len(t.torrent.KnownSwarm()),
	}
}

func (t *Torrent) countSeeders() int {
	count := 0
	for _, conn := range t.torrent.PeerConns() {
		numPieces := t.torrent.NumPieces()
		if numPieces >= 0 && conn.PeerPieces().GetCardinality() == uint64(numPieces) {
			count++
		}
	}
	return count
}

// File represents a file in a torrent.
type File struct {
	Path   string
	Length int64
}

// Stats represents torrent statistics.
type Stats struct {
	BytesRead         int64
	BytesWritten      int64
	BytesReadData     int64
	BytesWrittenData  int64
	ChunksReadWasted  int64
	ChunksWritten     int64
	PiecesDirtiedGood int64
	PiecesDirtiedBad  int64
	ActivePeers       int
	ConnectedSeeders  int
	TotalPeers        int
}

// String returns a formatted string of the stats.
func (s Stats) String() string {
	return fmt.Sprintf(
		"Read: %d bytes, Written: %d bytes, Peers: %d/%d, Seeders: %d",
		s.BytesReadData,
		s.BytesWrittenData,
		s.ActivePeers,
		s.TotalPeers,
		s.ConnectedSeeders,
	)
}
