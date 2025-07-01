package torrent

import (
	"context"
	"time"

	"github.com/ayutaz/orochi/internal/database"
	"github.com/ayutaz/orochi/internal/logger"
)

// ProgressUpdater periodically updates torrent progress in the database.
type ProgressUpdater struct {
	adapter  *ClientAdapter
	db       *database.DB
	logger   logger.Logger
	interval time.Duration
	cancel   context.CancelFunc
}

// NewProgressUpdater creates a new progress updater.
func NewProgressUpdater(adapter *ClientAdapter, db *database.DB, log logger.Logger) *ProgressUpdater {
	return &ProgressUpdater{
		adapter:  adapter,
		db:       db,
		logger:   log,
		interval: 5 * time.Second, // Update every 5 seconds
	}
}

// Start starts the progress updater.
func (u *ProgressUpdater) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	u.cancel = cancel

	go u.run(ctx)
}

// Stop stops the progress updater.
func (u *ProgressUpdater) Stop() {
	if u.cancel != nil {
		u.cancel()
	}
}

// run is the main loop for updating progress.
func (u *ProgressUpdater) run(ctx context.Context) {
	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.updateAll()
		}
	}
}

// updateAll updates progress for all active torrents.
func (u *ProgressUpdater) updateAll() {
	torrents := u.adapter.ListTorrents()

	for _, torrent := range torrents {
		// Only update if downloading or seeding
		if torrent.Status != StatusDownloading && torrent.Status != StatusSeeding {
			continue
		}

		// Update progress in database
		err := u.db.UpdateTorrentProgress(
			torrent.ID,
			torrent.Progress,
			torrent.Downloaded,
			torrent.Uploaded,
		)
		if err != nil {
			u.logger.Error("failed to update torrent progress",
				logger.String("id", torrent.ID),
				logger.Err(err),
			)
			continue
		}

		// Update status if needed
		if torrent.Status == StatusDownloading && torrent.Progress >= 100 {
			// Mark as completed
			if err := u.db.MarkTorrentCompleted(torrent.ID); err != nil {
				u.logger.Error("failed to mark torrent completed",
					logger.String("id", torrent.ID),
					logger.Err(err),
				)
			} else {
				u.logger.Info("torrent completed",
					logger.String("id", torrent.ID),
					logger.String("name", torrent.Info.Name),
				)
			}
		} else {
			// Update status
			if err := u.db.UpdateTorrentStatus(torrent.ID, string(torrent.Status)); err != nil {
				u.logger.Error("failed to update torrent status",
					logger.String("id", torrent.ID),
					logger.Err(err),
				)
			}
		}
	}
}
