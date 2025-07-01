package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/zeebo/bencode"
)

// TorrentInfo represents parsed torrent information.
type TorrentInfo struct {
	Name        string
	InfoHash    string
	Length      int64
	PieceLength int64
	Announce    string
	Trackers    []string
	Files       []FileInfo
}

// FileInfo represents a file in the torrent.
type FileInfo struct {
	Path   []string
	Length int64
}

// bencodeInfo represents the info dictionary from a torrent file.
type bencodeInfo struct {
	Name        string `bencode:"name"`
	PieceLength int64  `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Length      int64  `bencode:"length,omitempty"`
	Files       []struct {
		Length int64    `bencode:"length"`
		Path   []string `bencode:"path"`
	} `bencode:"files,omitempty"`
}

// bencodeTorrent represents the structure of a torrent file.
type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// ParseTorrentFile parses a torrent file and returns its information.
func ParseTorrentFile(data []byte) (*TorrentInfo, error) {
	var torrent bencodeTorrent
	
	if err := bencode.DecodeBytes(data, &torrent); err != nil {
		return nil, fmt.Errorf("failed to decode torrent: %w", err)
	}
	
	// Validate required fields
	if torrent.Info.Name == "" {
		return nil, errors.New("torrent missing required field: name")
	}
	if torrent.Info.PieceLength == 0 {
		return nil, errors.New("torrent missing required field: piece length")
	}
	if torrent.Info.Pieces == "" {
		return nil, errors.New("torrent missing required field: pieces")
	}
	
	// Calculate info hash
	infoBencode, err := bencode.EncodeBytes(torrent.Info)
	if err != nil {
		return nil, fmt.Errorf("failed to encode info dict: %w", err)
	}
	
	h := sha1.New()
	h.Write(infoBencode)
	infoHash := hex.EncodeToString(h.Sum(nil))
	
	info := &TorrentInfo{
		Name:        torrent.Info.Name,
		InfoHash:    infoHash,
		Length:      torrent.Info.Length,
		PieceLength: torrent.Info.PieceLength,
		Announce:    torrent.Announce,
		Trackers:    []string{torrent.Announce},
	}
	
	// Handle multi-file torrents
	if len(torrent.Info.Files) > 0 {
		var totalLength int64
		for _, f := range torrent.Info.Files {
			info.Files = append(info.Files, FileInfo{
				Path:   f.Path,
				Length: f.Length,
			})
			totalLength += f.Length
		}
		info.Length = totalLength
	}
	
	return info, nil
}

// ParseMagnetLink parses a magnet link and returns its information.
func ParseMagnetLink(magnetLink string) (*TorrentInfo, error) {
	if !strings.HasPrefix(magnetLink, "magnet:?") {
		return nil, errors.New("invalid magnet link format")
	}
	
	// Parse query parameters
	u, err := url.Parse(magnetLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse magnet link: %w", err)
	}
	
	params := u.Query()
	
	// Extract info hash (required)
	xt := params.Get("xt")
	if xt == "" {
		return nil, errors.New("magnet link missing required parameter: xt")
	}
	
	// Parse xt parameter (e.g., "urn:btih:1234567890abcdef...")
	if !strings.HasPrefix(xt, "urn:btih:") {
		return nil, errors.New("invalid xt parameter format")
	}
	
	infoHash := strings.TrimPrefix(xt, "urn:btih:")
	if len(infoHash) != 40 && len(infoHash) != 32 {
		return nil, errors.New("invalid info hash length")
	}
	
	// Convert base32 to hex if necessary
	if len(infoHash) == 32 {
		// TODO: Implement base32 to hex conversion
		return nil, errors.New("base32 info hash not yet supported")
	}
	
	info := &TorrentInfo{
		InfoHash: strings.ToLower(infoHash),
		Name:     params.Get("dn"), // Display name (optional)
		Trackers: params["tr"],      // Tracker URLs (optional)
	}
	
	// Parse optional parameters
	if xl := params.Get("xl"); xl != "" {
		if length, err := strconv.ParseInt(xl, 10, 64); err == nil {
			info.Length = length
		}
	}
	
	return info, nil
}