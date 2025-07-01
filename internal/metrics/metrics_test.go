package metrics

import (
	"testing"
	"time"
)

func TestMetrics_Torrents(t *testing.T) {
	m := &Metrics{
		HTTPRequestDuration: &DurationMetric{
			min: int64(time.Hour),
		},
		StartTime: time.Now(),
	}
	
	// Test increment/decrement
	m.IncrementTorrents()
	if m.TorrentsTotal != 1 {
		t.Errorf("expected TorrentsTotal 1, got %d", m.TorrentsTotal)
	}
	
	m.IncrementTorrents()
	if m.TorrentsTotal != 2 {
		t.Errorf("expected TorrentsTotal 2, got %d", m.TorrentsTotal)
	}
	
	m.DecrementTorrents()
	if m.TorrentsTotal != 1 {
		t.Errorf("expected TorrentsTotal 1, got %d", m.TorrentsTotal)
	}
}

func TestMetrics_SetTorrentStatus(t *testing.T) {
	m := &Metrics{
		HTTPRequestDuration: &DurationMetric{
			min: int64(time.Hour),
		},
		StartTime: time.Now(),
	}
	
	// Test status transitions
	m.SetTorrentStatus("", "stopped")
	if m.TorrentsStopped != 1 {
		t.Errorf("expected TorrentsStopped 1, got %d", m.TorrentsStopped)
	}
	
	m.SetTorrentStatus("stopped", "downloading")
	if m.TorrentsStopped != 0 {
		t.Errorf("expected TorrentsStopped 0, got %d", m.TorrentsStopped)
	}
	if m.TorrentsDownloading != 1 {
		t.Errorf("expected TorrentsDownloading 1, got %d", m.TorrentsDownloading)
	}
	
	m.SetTorrentStatus("downloading", "seeding")
	if m.TorrentsDownloading != 0 {
		t.Errorf("expected TorrentsDownloading 0, got %d", m.TorrentsDownloading)
	}
	if m.TorrentsSeeding != 1 {
		t.Errorf("expected TorrentsSeeding 1, got %d", m.TorrentsSeeding)
	}
	
	m.SetTorrentStatus("seeding", "error")
	if m.TorrentsSeeding != 0 {
		t.Errorf("expected TorrentsSeeding 0, got %d", m.TorrentsSeeding)
	}
	if m.TorrentsError != 1 {
		t.Errorf("expected TorrentsError 1, got %d", m.TorrentsError)
	}
}

func TestMetrics_Transfer(t *testing.T) {
	m := &Metrics{
		HTTPRequestDuration: &DurationMetric{
			min: int64(time.Hour),
		},
		StartTime: time.Now(),
	}
	
	m.AddBytesDownloaded(1024)
	if m.BytesDownloaded != 1024 {
		t.Errorf("expected BytesDownloaded 1024, got %d", m.BytesDownloaded)
	}
	
	m.AddBytesDownloaded(2048)
	if m.BytesDownloaded != 3072 {
		t.Errorf("expected BytesDownloaded 3072, got %d", m.BytesDownloaded)
	}
	
	m.AddBytesUploaded(512)
	if m.BytesUploaded != 512 {
		t.Errorf("expected BytesUploaded 512, got %d", m.BytesUploaded)
	}
}

func TestMetrics_HTTP(t *testing.T) {
	m := &Metrics{
		HTTPRequestDuration: &DurationMetric{
			min: int64(time.Hour),
		},
		StartTime: time.Now(),
	}
	
	m.IncrementHTTPRequests()
	m.IncrementHTTPRequests()
	if m.HTTPRequests != 2 {
		t.Errorf("expected HTTPRequests 2, got %d", m.HTTPRequests)
	}
	
	m.IncrementHTTPErrors()
	if m.HTTPErrors != 1 {
		t.Errorf("expected HTTPErrors 1, got %d", m.HTTPErrors)
	}
}

func TestDurationMetric(t *testing.T) {
	d := &DurationMetric{
		min: int64(time.Hour),
	}
	
	// Record some durations
	d.Record(100 * time.Millisecond)
	d.Record(200 * time.Millisecond)
	d.Record(50 * time.Millisecond)
	d.Record(300 * time.Millisecond)
	
	count, avg, min, max := d.Stats()
	
	if count != 4 {
		t.Errorf("expected count 4, got %d", count)
	}
	
	expectedAvg := 162500000 * time.Nanosecond // (100+200+50+300)/4 = 162.5ms
	if avg != expectedAvg {
		t.Errorf("expected avg %v, got %v", expectedAvg, avg)
	}
	
	if min != 50*time.Millisecond {
		t.Errorf("expected min 50ms, got %v", min)
	}
	
	if max != 300*time.Millisecond {
		t.Errorf("expected max 300ms, got %v", max)
	}
}

func TestMetrics_Snapshot(t *testing.T) {
	m := &Metrics{
		HTTPRequestDuration: &DurationMetric{
			min: int64(time.Hour),
		},
		StartTime: time.Now(),
	}
	
	// Set some metrics
	m.TorrentsTotal = 5
	m.TorrentsDownloading = 2
	m.TorrentsSeeding = 1
	m.TorrentsStopped = 2
	m.BytesDownloaded = 1024 * 1024
	m.BytesUploaded = 512 * 1024
	m.HTTPRequests = 100
	m.HTTPErrors = 5
	m.SetMemoryUsage(64 * 1024 * 1024)
	m.SetGoroutineCount(50)
	
	// Record some HTTP durations
	m.RecordHTTPDuration(100 * time.Millisecond)
	m.RecordHTTPDuration(200 * time.Millisecond)
	
	snapshot := m.Snapshot()
	
	// Check torrent metrics
	torrents := snapshot["torrents"].(map[string]int64)
	if torrents["total"] != 5 {
		t.Errorf("expected total 5, got %d", torrents["total"])
	}
	if torrents["downloading"] != 2 {
		t.Errorf("expected downloading 2, got %d", torrents["downloading"])
	}
	
	// Check transfer metrics
	transfer := snapshot["transfer"].(map[string]int64)
	if transfer["downloaded"] != 1024*1024 {
		t.Errorf("expected downloaded %d, got %d", 1024*1024, transfer["downloaded"])
	}
	
	// Check HTTP metrics
	http := snapshot["http"].(map[string]interface{})
	if http["requests"].(int64) != 100 {
		t.Errorf("expected requests 100, got %d", http["requests"])
	}
	
	// Check system metrics
	system := snapshot["system"].(map[string]interface{})
	if system["memory_bytes"].(int64) != 64*1024*1024 {
		t.Errorf("expected memory_bytes %d, got %d", 64*1024*1024, system["memory_bytes"])
	}
	if system["goroutines"].(int32) != 50 {
		t.Errorf("expected goroutines 50, got %d", system["goroutines"])
	}
}

func TestGet_Singleton(t *testing.T) {
	// Get should return the same instance
	m1 := Get()
	m2 := Get()
	
	if m1 != m2 {
		t.Error("Get() should return the same instance")
	}
	
	// Verify it's properly initialized
	if m1.HTTPRequestDuration == nil {
		t.Error("HTTPRequestDuration should be initialized")
	}
	
	if m1.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}