package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// DB represents the database connection.
type DB struct {
	db     *sql.DB
	logger logger.Logger
}

// TorrentRecord represents a torrent record in the database.
type TorrentRecord struct {
	ID           string     `json:"id"`
	InfoHash     string     `json:"info_hash"`
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	Downloaded   int64      `json:"downloaded"`
	Uploaded     int64      `json:"uploaded"`
	DownloadPath string     `json:"download_path"`
	AddedAt      time.Time  `json:"added_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Metadata     string     `json:"metadata"` // JSON encoded torrent file data
}

// NewDB creates a new database connection.
func NewDB(dbPath string, log logger.Logger) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.InternalErrorf("failed to open database: %v", err)
	}

	// Configure connection pool for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, errors.InternalErrorf("failed to ping database: %v", err)
	}

	d := &DB{
		db:     db,
		logger: log,
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Warn("failed to enable WAL mode", logger.Err(err))
	}

	// Other performance optimizations
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		log.Warn("failed to set synchronous mode", logger.Err(err))
	}

	if _, err := db.Exec("PRAGMA cache_size=10000"); err != nil {
		log.Warn("failed to set cache size", logger.Err(err))
	}

	if err := d.createTables(); err != nil {
		return nil, err
	}

	return d, nil
}

// createTables creates the necessary database tables.
func (d *DB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS torrents (
		id TEXT PRIMARY KEY,
		info_hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		status TEXT NOT NULL,
		progress REAL NOT NULL DEFAULT 0,
		downloaded INTEGER NOT NULL DEFAULT 0,
		uploaded INTEGER NOT NULL DEFAULT 0,
		download_path TEXT NOT NULL,
		added_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		metadata TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_torrents_status ON torrents(status);
	CREATE INDEX IF NOT EXISTS idx_torrents_added_at ON torrents(added_at);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return errors.InternalErrorf("failed to create tables: %v", err)
	}

	return nil
}

// SaveTorrent saves a torrent record to the database.
func (d *DB) SaveTorrent(record *TorrentRecord) error {
	query := `
	INSERT OR REPLACE INTO torrents (
		id, info_hash, name, size, status, progress, 
		downloaded, uploaded, download_path, added_at, 
		completed_at, metadata, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query,
		record.ID,
		record.InfoHash,
		record.Name,
		record.Size,
		record.Status,
		record.Progress,
		record.Downloaded,
		record.Uploaded,
		record.DownloadPath,
		record.AddedAt,
		record.CompletedAt,
		record.Metadata,
	)
	if err != nil {
		return errors.InternalErrorf("failed to save torrent: %v", err)
	}

	return nil
}

// GetTorrent retrieves a torrent record by ID.
func (d *DB) GetTorrent(id string) (*TorrentRecord, error) {
	query := `
	SELECT id, info_hash, name, size, status, progress, 
		downloaded, uploaded, download_path, added_at, 
		completed_at, metadata
	FROM torrents
	WHERE id = ?
	`

	var record TorrentRecord
	var completedAt sql.NullTime

	err := d.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.InfoHash,
		&record.Name,
		&record.Size,
		&record.Status,
		&record.Progress,
		&record.Downloaded,
		&record.Uploaded,
		&record.DownloadPath,
		&record.AddedAt,
		&completedAt,
		&record.Metadata,
	)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("torrent %s not found", id)
	}
	if err != nil {
		return nil, errors.InternalErrorf("failed to get torrent: %v", err)
	}

	if completedAt.Valid {
		record.CompletedAt = &completedAt.Time
	}

	return &record, nil
}

// ListTorrents retrieves all torrent records.
func (d *DB) ListTorrents() ([]*TorrentRecord, error) {
	query := `
	SELECT id, info_hash, name, size, status, progress, 
		downloaded, uploaded, download_path, added_at, 
		completed_at, metadata
	FROM torrents
	ORDER BY added_at DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, errors.InternalErrorf("failed to list torrents: %v", err)
	}
	defer rows.Close()

	var records []*TorrentRecord
	for rows.Next() {
		var record TorrentRecord
		var completedAt sql.NullTime

		err := rows.Scan(
			&record.ID,
			&record.InfoHash,
			&record.Name,
			&record.Size,
			&record.Status,
			&record.Progress,
			&record.Downloaded,
			&record.Uploaded,
			&record.DownloadPath,
			&record.AddedAt,
			&completedAt,
			&record.Metadata,
		)
		if err != nil {
			return nil, errors.InternalErrorf("failed to scan torrent: %v", err)
		}

		if completedAt.Valid {
			record.CompletedAt = &completedAt.Time
		}

		records = append(records, &record)
	}

	return records, nil
}

// DeleteTorrent deletes a torrent record by ID.
func (d *DB) DeleteTorrent(id string) error {
	query := `DELETE FROM torrents WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return errors.InternalErrorf("failed to delete torrent: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
	}

	return nil
}

// UpdateTorrentProgress updates the progress of a torrent.
func (d *DB) UpdateTorrentProgress(id string, progress float64, downloaded, uploaded int64) error {
	query := `
	UPDATE torrents 
	SET progress = ?, downloaded = ?, uploaded = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	result, err := d.db.Exec(query, progress, downloaded, uploaded, id)
	if err != nil {
		return errors.InternalErrorf("failed to update torrent progress: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
	}

	return nil
}

// UpdateTorrentStatus updates the status of a torrent.
func (d *DB) UpdateTorrentStatus(id, status string) error {
	query := `
	UPDATE torrents 
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	result, err := d.db.Exec(query, status, id)
	if err != nil {
		return errors.InternalErrorf("failed to update torrent status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
	}

	return nil
}

// ProgressUpdate represents a batch progress update.
type ProgressUpdate struct {
	ID         string
	Progress   float64
	Downloaded int64
	Uploaded   int64
}

// UpdateTorrentProgressBatch updates progress for multiple torrents in a single transaction.
func (d *DB) UpdateTorrentProgressBatch(updates []ProgressUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return errors.InternalErrorf("failed to begin transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(`
		UPDATE torrents 
		SET progress = ?, downloaded = ?, uploaded = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`)
	if err != nil {
		return errors.InternalErrorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, update := range updates {
		_, err = stmt.Exec(update.Progress, update.Downloaded, update.Uploaded, update.ID)
		if err != nil {
			return errors.InternalErrorf("failed to update torrent %s: %v", update.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.InternalErrorf("failed to commit transaction: %v", err)
	}

	return nil
}

// MarkTorrentCompleted marks a torrent as completed.
func (d *DB) MarkTorrentCompleted(id string) error {
	query := `
	UPDATE torrents 
	SET status = 'completed', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return errors.InternalErrorf("failed to mark torrent completed: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %v", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
	}

	return nil
}

// SaveSetting saves a setting to the database.
func (d *DB) SaveSetting(key, value string) error {
	query := `
	INSERT OR REPLACE INTO settings (key, value, updated_at) 
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query, key, value)
	if err != nil {
		return errors.InternalErrorf("failed to save setting: %v", err)
	}

	return nil
}

// GetSetting retrieves a setting by key.
func (d *DB) GetSetting(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?`

	var value string
	err := d.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", errors.NotFoundf("setting %s not found", key)
	}
	if err != nil {
		return "", errors.InternalErrorf("failed to get setting: %v", err)
	}

	return value, nil
}

// GetAllSettings retrieves all settings as a map.
func (d *DB) GetAllSettings() (map[string]string, error) {
	query := `SELECT key, value FROM settings`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, errors.InternalErrorf("failed to get all settings: %v", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, errors.InternalErrorf("failed to scan setting: %v", err)
		}
		settings[key] = value
	}

	return settings, nil
}

// SaveSettingsJSON saves settings as JSON.
func (d *DB) SaveSettingsJSON(settings interface{}) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return errors.InternalErrorf("failed to marshal settings: %v", err)
	}

	return d.SaveSetting("app_settings", string(data))
}

// GetSettingsJSON retrieves settings as JSON.
func (d *DB) GetSettingsJSON(settings interface{}) error {
	data, err := d.GetSetting("app_settings")
	if err != nil {
		if errors.IsNotFound(err) {
			return nil // No settings yet
		}
		return err
	}

	if err := json.Unmarshal([]byte(data), settings); err != nil {
		return errors.InternalErrorf("failed to unmarshal settings: %v", err)
	}

	return nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}
