package torrent

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	t.Run("マネージャーを作成できる", func(t *testing.T) {
		manager := NewManager()
		
		if manager == nil {
			t.Fatal("manager should not be nil")
		}
		
		// Check initialization by behavior, not internals
		if manager.Count() != 0 {
			t.Errorf("expected 0 torrents, got %d", manager.Count())
		}
	})
}

func TestManager_AddTorrent(t *testing.T) {
	t.Run("torrentファイルからトレントを追加できる", func(t *testing.T) {
		manager := NewManager()
		torrentData := CreateTestTorrent()
		
		id, err := manager.AddTorrent(torrentData)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if id == "" {
			t.Error("torrent ID should not be empty")
		}
		
		if manager.Count() != 1 {
			t.Errorf("expected 1 torrent, got %d", manager.Count())
		}
		
		// Verify torrent exists
		torrent, exists := manager.GetTorrent(id)
		if !exists {
			t.Error("torrent should exist")
		}
		
		if torrent.Info.Name != "test.txt" {
			t.Errorf("expected name 'test.txt', got %s", torrent.Info.Name)
		}
	})
	
	t.Run("同じトレントを二回追加すると同じIDを返す", func(t *testing.T) {
		manager := NewManager()
		torrentData := CreateTestTorrent()
		
		id1, err1 := manager.AddTorrent(torrentData)
		if err1 != nil {
			t.Fatalf("first add failed: %v", err1)
		}
		
		id2, err2 := manager.AddTorrent(torrentData)
		if err2 != nil {
			t.Fatalf("second add failed: %v", err2)
		}
		
		if id1 != id2 {
			t.Errorf("expected same ID for same torrent, got %s and %s", id1, id2)
		}
		
		if manager.Count() != 1 {
			t.Errorf("expected 1 torrent, got %d", manager.Count())
		}
	})
	
	t.Run("不正なトレントデータはエラーを返す", func(t *testing.T) {
		manager := NewManager()
		invalidData := []byte("invalid torrent data")
		
		_, err := manager.AddTorrent(invalidData)
		
		if err == nil {
			t.Error("expected error for invalid torrent data")
		}
		
		if manager.Count() != 0 {
			t.Errorf("expected 0 torrents after failed add, got %d", manager.Count())
		}
	})
}

func TestManager_AddMagnet(t *testing.T) {
	t.Run("マグネットリンクからトレントを追加できる", func(t *testing.T) {
		manager := NewManager()
		magnetLink := "magnet:?xt=urn:btih:1234567890abcdef1234567890abcdef12345678&dn=test.txt"
		
		id, err := manager.AddMagnet(magnetLink)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if id != "1234567890abcdef1234567890abcdef12345678" {
			t.Errorf("expected ID to match info hash, got %s", id)
		}
		
		if manager.Count() != 1 {
			t.Errorf("expected 1 torrent, got %d", manager.Count())
		}
	})
}

func TestManager_RemoveTorrent(t *testing.T) {
	t.Run("トレントを削除できる", func(t *testing.T) {
		manager := NewManager()
		torrentData := CreateTestTorrent()
		
		id, _ := manager.AddTorrent(torrentData)
		
		err := manager.RemoveTorrent(id)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if manager.Count() != 0 {
			t.Errorf("expected 0 torrents after removal, got %d", manager.Count())
		}
		
		_, exists := manager.GetTorrent(id)
		if exists {
			t.Error("torrent should not exist after removal")
		}
	})
	
	t.Run("存在しないトレントの削除はエラーを返す", func(t *testing.T) {
		manager := NewManager()
		
		err := manager.RemoveTorrent("nonexistent")
		
		if err == nil {
			t.Error("expected error for nonexistent torrent")
		}
	})
}

func TestManager_StartTorrent(t *testing.T) {
	t.Run("トレントを開始できる", func(t *testing.T) {
		manager := NewManager()
		torrentData := CreateTestTorrent()
		
		id, _ := manager.AddTorrent(torrentData)
		
		err := manager.StartTorrent(id)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		torrent, _ := manager.GetTorrent(id)
		if torrent.Status != StatusDownloading {
			t.Errorf("expected status %s, got %s", StatusDownloading, torrent.Status)
		}
	})
}

func TestManager_StopTorrent(t *testing.T) {
	t.Run("トレントを停止できる", func(t *testing.T) {
		manager := NewManager()
		torrentData := CreateTestTorrent()
		
		id, _ := manager.AddTorrent(torrentData)
		if err := manager.StartTorrent(id); err != nil {
			t.Errorf("failed to start torrent: %v", err)
		}
		
		err := manager.StopTorrent(id)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		torrent, _ := manager.GetTorrent(id)
		if torrent.Status != StatusStopped {
			t.Errorf("expected status %s, got %s", StatusStopped, torrent.Status)
		}
	})
}

func TestManager_ListTorrents(t *testing.T) {
	t.Run("すべてのトレントをリストできる", func(t *testing.T) {
		manager := NewManager()
		
		// Add multiple torrents
		torrentData := CreateTestTorrent()
		id1, _ := manager.AddTorrent(torrentData)
		
		magnetLink := "magnet:?xt=urn:btih:abcdef1234567890abcdef1234567890abcdef12&dn=test2.txt"
		id2, err := manager.AddMagnet(magnetLink)
		if err != nil {
			t.Fatalf("failed to add magnet: %v", err)
		}
		
		torrents := manager.ListTorrents()
		
		if len(torrents) != 2 {
			t.Fatalf("expected 2 torrents, got %d", len(torrents))
		}
		
		// Check that both torrents are in the list
		foundID1 := false
		foundID2 := false
		for _, torrent := range torrents {
			if torrent.ID == id1 {
				foundID1 = true
			}
			if torrent.ID == id2 {
				foundID2 = true
			}
		}
		
		if !foundID1 || !foundID2 {
			t.Error("not all torrents were returned in the list")
		}
	})
}

func TestTorrent_UpdateProgress(t *testing.T) {
	t.Run("進捗を更新できる", func(t *testing.T) {
		torrent := &Torrent{
			Info: &TorrentInfo{
				Length: 1000,
			},
			Downloaded: 0,
			AddedAt:    time.Now(),
		}
		
		torrent.UpdateProgress(500, 200)
		
		if torrent.Downloaded != 500 {
			t.Errorf("expected downloaded 500, got %d", torrent.Downloaded)
		}
		
		if torrent.Uploaded != 200 {
			t.Errorf("expected uploaded 200, got %d", torrent.Uploaded)
		}
		
		expectedProgress := 50.0
		if torrent.Progress != expectedProgress {
			t.Errorf("expected progress %.1f%%, got %.1f%%", expectedProgress, torrent.Progress)
		}
	})
}