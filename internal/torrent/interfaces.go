package torrent

// Manager defines the interface for torrent management operations.
type Manager interface {
	AddTorrent(data []byte) (string, error)
	AddMagnet(magnetLink string) (string, error)
	RemoveTorrent(id string) error
	GetTorrent(id string) (*Torrent, bool)
	ListTorrents() []*Torrent
	StartTorrent(id string) error
	StopTorrent(id string) error
	Count() int
}

// Parser defines the interface for parsing torrent files and magnet links.
type Parser interface {
	ParseTorrentFile(data []byte) (*TorrentInfo, error)
	ParseMagnetLink(magnetLink string) (*TorrentInfo, error)
}

// Storage defines the interface for persistent storage operations.
type Storage interface {
	Save(torrent *Torrent) error
	Load(id string) (*Torrent, error)
	Delete(id string) error
	List() ([]*Torrent, error)
}

// Downloader defines the interface for torrent download operations.
type Downloader interface {
	Start(torrent *Torrent) error
	Stop(torrent *Torrent) error
	GetProgress(torrent *Torrent) (downloaded, uploaded int64, err error)
}

// Tracker defines the interface for tracker communication.
type Tracker interface {
	Announce(torrent *Torrent) ([]Peer, error)
	Scrape(torrent *Torrent) (*ScrapeResponse, error)
}

// Peer represents a peer in the swarm.
type Peer struct {
	IP   string
	Port int
}

// ScrapeResponse represents a tracker scrape response.
type ScrapeResponse struct {
	Seeders   int
	Leechers  int
	Completed int
}
