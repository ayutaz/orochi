package torrent

import "github.com/zeebo/bencode"

// CreateTestTorrent creates a valid test torrent data for testing.
func CreateTestTorrent() []byte {
	torrent := bencodeTorrent{
		Announce: "http://example.com:8000",
		Info: bencodeInfo{
			Name:        "test.txt",
			PieceLength: 16384,
			Length:      1024,
			Pieces:      "01234567890123456789",
		},
	}

	data, err := bencode.EncodeBytes(torrent)
	if err != nil {
		panic(err) // This should never happen in tests
	}

	return data
}
