package torrent

import (
	"testing"
)

func TestParseTorrentFile(t *testing.T) {
	t.Run("有効なtorrentファイルをパースできる", func(t *testing.T) {
		torrentData := CreateTestTorrent()

		info, err := ParseTorrentFile(torrentData)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info.Name != "test.txt" {
			t.Errorf("expected name 'test.txt', got %s", info.Name)
		}

		if info.Length != 1024 {
			t.Errorf("expected length 1024, got %d", info.Length)
		}

		if info.PieceLength != 16384 {
			t.Errorf("expected piece length 16384, got %d", info.PieceLength)
		}

		if info.Announce != "http://example.com:8000" {
			t.Errorf("expected announce 'http://example.com:8000', got %s", info.Announce)
		}

		if info.InfoHash == "" {
			t.Error("info hash should not be empty")
		}
	})

	t.Run("不正なtorrentファイルはエラーを返す", func(t *testing.T) {
		torrentData := []byte("invalid data")

		_, err := ParseTorrentFile(torrentData)

		if err == nil {
			t.Error("expected error for invalid torrent data")
		}
	})

	t.Run("必須フィールドが欠けている場合はエラーを返す", func(t *testing.T) {
		// Missing 'info' dictionary
		torrentData := []byte("d8:announce21:http://example.com:8000e")

		_, err := ParseTorrentFile(torrentData)

		if err == nil {
			t.Error("expected error for missing info field")
		}
	})
}

func TestParseMagnetLink(t *testing.T) {
	t.Run("有効なマグネットリンクをパースできる", func(t *testing.T) {
		magnetLink := "magnet:?xt=urn:btih:1234567890abcdef1234567890abcdef12345678&" +
			"dn=test.txt&tr=http://tracker.example.com:8080/announce"

		info, err := ParseMagnetLink(magnetLink)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info.InfoHash != "1234567890abcdef1234567890abcdef12345678" {
			t.Errorf("expected info hash '1234567890abcdef1234567890abcdef12345678', got %s", info.InfoHash)
		}

		if info.Name != "test.txt" {
			t.Errorf("expected name 'test.txt', got %s", info.Name)
		}

		if len(info.Trackers) != 1 || info.Trackers[0] != "http://tracker.example.com:8080/announce" {
			t.Errorf("expected tracker 'http://tracker.example.com:8080/announce', got %v", info.Trackers)
		}
	})

	t.Run("必須パラメータがない場合はエラーを返す", func(t *testing.T) {
		magnetLink := "magnet:?dn=test.txt" // Missing xt parameter

		_, err := ParseMagnetLink(magnetLink)

		if err == nil {
			t.Error("expected error for missing xt parameter")
		}
	})

	t.Run("無効な形式の場合はエラーを返す", func(t *testing.T) {
		magnetLink := "http://example.com/file.torrent" // Not a magnet link

		_, err := ParseMagnetLink(magnetLink)

		if err == nil {
			t.Error("expected error for invalid magnet link format")
		}
	})
}
