package torrent

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ayutaz/orochi/internal/errors"
)

// TestManagerIntegration tests the integration between components.
func TestManagerIntegration(t *testing.T) {
	t.Run("複数のトレントを同時に管理できる", func(t *testing.T) {
		manager := NewManager()

		// Add multiple torrents
		var ids []string
		for i := 0; i < 10; i++ {
			data := createTestTorrentWithName(fmt.Sprintf("test%d.txt", i))
			id, err := manager.AddTorrent(data)
			if err != nil {
				t.Fatalf("failed to add torrent %d: %v", i, err)
			}
			ids = append(ids, id)
		}

		// Verify all torrents exist
		if manager.Count() != 10 {
			t.Errorf("expected 10 torrents, got %d", manager.Count())
		}

		// Start some torrents
		for i := 0; i < 5; i++ {
			if err := manager.StartTorrent(ids[i]); err != nil {
				t.Errorf("failed to start torrent %s: %v", ids[i], err)
			}
		}

		// Verify status changes
		for i := 0; i < 5; i++ {
			torrent, exists := manager.GetTorrent(ids[i])
			if !exists {
				t.Errorf("torrent %s not found", ids[i])
				continue
			}
			if torrent.Status != StatusDownloading {
				t.Errorf("torrent %s status should be downloading, got %s", ids[i], torrent.Status)
			}
		}

		// Stop some torrents
		for i := 0; i < 3; i++ {
			if err := manager.StopTorrent(ids[i]); err != nil {
				t.Errorf("failed to stop torrent %s: %v", ids[i], err)
			}
		}

		// Remove some torrents
		for i := 7; i < 10; i++ {
			if err := manager.RemoveTorrent(ids[i]); err != nil {
				t.Errorf("failed to remove torrent %s: %v", ids[i], err)
			}
		}

		// Final verification
		if manager.Count() != 7 {
			t.Errorf("expected 7 torrents after removal, got %d", manager.Count())
		}
	})

	t.Run("並行アクセスで競合状態が発生しない", func(t *testing.T) {
		manager := NewManager()

		// Number of goroutines and operations
		numGoroutines := 50
		numOperations := 100

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numOperations)

		// Launch goroutines for concurrent operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < numOperations; j++ {
					switch j % 5 {
					case 0:
						// Add torrent
						data := createTestTorrentWithName(fmt.Sprintf("routine%d-op%d.txt", routineID, j))
						if _, err := manager.AddTorrent(data); err != nil {
							errors <- err
						}
					case 1:
						// Add magnet
						magnet := fmt.Sprintf("magnet:?xt=urn:btih:%040d&dn=test%d", routineID*1000+j, j)
						if _, err := manager.AddMagnet(magnet); err != nil {
							errors <- err
						}
					case 2:
						// List torrents
						torrents := manager.ListTorrents()
						_ = torrents // Just access it
					case 3:
						// Count torrents
						count := manager.Count()
						_ = count // Just access it
					case 4:
						// Get random torrent
						manager.GetTorrent("somekey")
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errors)

		// Check for errors
		var errorCount int
		for err := range errors {
			if err != nil {
				errorCount++
				t.Logf("concurrent operation error: %v", err)
			}
		}

		if errorCount > 0 {
			t.Errorf("encountered %d errors during concurrent operations", errorCount)
		}

		// Verify manager is still functional
		finalCount := manager.Count()
		t.Logf("Final torrent count: %d", finalCount)

		// List should not panic
		torrents := manager.ListTorrents()
		if len(torrents) != finalCount {
			t.Errorf("list returned %d torrents, but count is %d", len(torrents), finalCount)
		}
	})

	t.Run("トレント情報の整合性が保たれる", func(t *testing.T) {
		manager := NewManager()

		// Add a torrent
		data := CreateTestTorrent()
		id, err := manager.AddTorrent(data)
		if err != nil {
			t.Fatalf("failed to add torrent: %v", err)
		}

		// Get the torrent
		torrent1, exists := manager.GetTorrent(id)
		if !exists {
			t.Fatal("torrent should exist")
		}

		// Modify status
		if err := manager.StartTorrent(id); err != nil {
			t.Fatalf("failed to start torrent: %v", err)
		}

		// Get again and verify
		torrent2, exists := manager.GetTorrent(id)
		if !exists {
			t.Fatal("torrent should still exist")
		}

		// Verify data integrity
		if torrent1.ID != torrent2.ID {
			t.Error("torrent ID changed")
		}
		if torrent1.Info.Name != torrent2.Info.Name {
			t.Error("torrent name changed")
		}
		if torrent1.Info.InfoHash != torrent2.Info.InfoHash {
			t.Error("torrent info hash changed")
		}

		// Status should have changed
		if torrent2.Status != StatusDownloading {
			t.Errorf("status should be downloading, got %s", torrent2.Status)
		}
	})
}

// TestErrorPropagation tests that errors are properly propagated.
func TestErrorPropagation(t *testing.T) {
	manager := NewManager()

	t.Run("不正なトレントデータのエラーが伝播する", func(t *testing.T) {
		invalidData := []byte("this is not a valid torrent")

		_, err := manager.AddTorrent(invalidData)
		if err == nil {
			t.Fatal("expected error for invalid torrent data")
		}

		// Verify error type
		if !errors.IsParseError(err) {
			t.Errorf("expected parse error, got %v", err)
		}
	})

	t.Run("不正なマグネットリンクのエラーが伝播する", func(t *testing.T) {
		invalidMagnet := "not-a-magnet-link"

		_, err := manager.AddMagnet(invalidMagnet)
		if err == nil {
			t.Fatal("expected error for invalid magnet link")
		}

		// Verify error type
		if !errors.IsInvalidInput(err) {
			t.Errorf("expected invalid input error, got %v", err)
		}
	})

	t.Run("存在しないトレントの操作でエラーが返る", func(t *testing.T) {
		nonExistentID := "non-existent-id"

		// Test RemoveTorrent
		err := manager.RemoveTorrent(nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent torrent")
		}
		if !errors.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}

		// Test StartTorrent
		err = manager.StartTorrent(nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent torrent")
		}
		if !errors.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}

		// Test StopTorrent
		err = manager.StopTorrent(nonExistentID)
		if err == nil {
			t.Fatal("expected error for non-existent torrent")
		}
		if !errors.IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

// TestPerformanceUnderLoad tests performance under high load.
func TestPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	manager := NewManager()

	// Add 1000 torrents
	start := time.Now()
	for i := 0; i < 1000; i++ {
		data := createTestTorrentWithName(fmt.Sprintf("load-test-%d.txt", i))
		if _, err := manager.AddTorrent(data); err != nil {
			t.Fatalf("failed to add torrent %d: %v", i, err)
		}
	}
	addDuration := time.Since(start)

	// List all torrents
	start = time.Now()
	torrents := manager.ListTorrents()
	listDuration := time.Since(start)

	if len(torrents) != 1000 {
		t.Errorf("expected 1000 torrents, got %d", len(torrents))
	}

	t.Logf("Added 1000 torrents in %v", addDuration)
	t.Logf("Listed 1000 torrents in %v", listDuration)

	// Performance expectations
	if addDuration > 1*time.Second {
		t.Errorf("adding 1000 torrents took too long: %v", addDuration)
	}
	if listDuration > 10*time.Millisecond {
		t.Errorf("listing 1000 torrents took too long: %v", listDuration)
	}
}
