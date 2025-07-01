package database

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ayutaz/orochi/internal/errors"
	"github.com/ayutaz/orochi/internal/logger"
)

// DB represents the database connection.
type DB struct {
	db     *sql.DB
	logger logger.Logger
}

// TorrentRecord represents a torrent record in the database.
type TorrentRecord struct {
	ID           string    `json:"id"`
	InfoHash     string    `json:"info_hash"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Status       string    `json:"status"`
	Progress     float64   `json:"progress"`
	Downloaded   int64     `json:"downloaded"`
	Uploaded     int64     `json:"uploaded"`
	DownloadPath string    `json:"download_path"`
	AddedAt      time.Time `json:"added_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Metadata     string    `json:"metadata"` // JSON encoded torrent file data
}

// NewDB creates a new database connection.
func NewDB(dbPath string, log logger.Logger) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.InternalErrorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, errors.InternalErrorf("failed to ping database: %w", err)
	}

	d := &DB{
		db:     db,
		logger: log,
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
		return errors.InternalErrorf("failed to create tables: %w", err)
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
		return errors.InternalErrorf("failed to save torrent: %w", err)
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
		return nil, errors.InternalErrorf("failed to get torrent: %w", err)
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
		return nil, errors.InternalErrorf("failed to list torrents: %w", err)
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
			return nil, errors.InternalErrorf("failed to scan torrent: %w", err)
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
		return errors.InternalErrorf("failed to delete torrent: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %w", err)
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
		return errors.InternalErrorf("failed to update torrent progress: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
	}

	return nil
}

// UpdateTorrentStatus updates the status of a torrent.
func (d *DB) UpdateTorrentStatus(id string, status string) error {
	query := `
	UPDATE torrents 
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	result, err := d.db.Exec(query, status, id)
	if err != nil {
		return errors.InternalErrorf("failed to update torrent status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.NotFoundf("torrent %s not found", id)
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
		return errors.InternalErrorf("failed to mark torrent completed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.InternalErrorf("failed to get rows affected: %w", err)
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
		return errors.InternalErrorf("failed to save setting: %w", err)
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
		return "", errors.InternalErrorf("failed to get setting: %w", err)
	}

	return value, nil
}

// GetAllSettings retrieves all settings as a map.
func (d *DB) GetAllSettings() (map[string]string, error) {
	query := `SELECT key, value FROM settings`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, errors.InternalErrorf("failed to get all settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, errors.InternalErrorf("failed to scan setting: %w", err)
		}
		settings[key] = value
	}

	return settings, nil
}

// SaveSettingsJSON saves settings as JSON.
func (d *DB) SaveSettingsJSON(settings interface{}) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return errors.InternalErrorf("failed to marshal settings: %w", err)
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
		return errors.InternalErrorf("failed to unmarshal settings: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}