package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds application metrics.
type Metrics struct {
	// Time tracking
	StartTime           time.Time
	HTTPRequestDuration *DurationMetric

	// Torrent metrics
	TorrentsTotal       int64
	TorrentsDownloading int64
	TorrentsSeeding     int64
	TorrentsStopped     int64
	TorrentsError       int64

	// Transfer metrics
	BytesDownloaded int64
	BytesUploaded   int64

	// HTTP metrics
	HTTPRequests int64
	HTTPErrors   int64
	MemoryUsage  int64

	// Performance metrics
	GoroutineCount int32
}

// DurationMetric tracks duration statistics.
type DurationMetric struct {
	mu    sync.RWMutex
	count int64
	sum   int64
	min   int64
	max   int64
}

// Global metrics instance.
var (
	globalMetrics *Metrics
	once          sync.Once
)

// Get returns the global metrics instance
func Get() *Metrics {
	once.Do(func() {
		globalMetrics = &Metrics{
			HTTPRequestDuration: &DurationMetric{
				min: int64(time.Hour), // Initialize with large value
			},
			StartTime: time.Now(),
		}
	})
	return globalMetrics
}

// IncrementTorrents increments the total torrent count
func (m *Metrics) IncrementTorrents() {
	atomic.AddInt64(&m.TorrentsTotal, 1)
}

// DecrementTorrents decrements the total torrent count
func (m *Metrics) DecrementTorrents() {
	atomic.AddInt64(&m.TorrentsTotal, -1)
}

// SetTorrentStatus updates torrent status counters
func (m *Metrics) SetTorrentStatus(oldStatus, newStatus string) {
	// Decrement old status
	switch oldStatus {
	case "downloading":
		atomic.AddInt64(&m.TorrentsDownloading, -1)
	case "seeding":
		atomic.AddInt64(&m.TorrentsSeeding, -1)
	case "stopped":
		atomic.AddInt64(&m.TorrentsStopped, -1)
	case "error":
		atomic.AddInt64(&m.TorrentsError, -1)
	}

	// Increment new status
	switch newStatus {
	case "downloading":
		atomic.AddInt64(&m.TorrentsDownloading, 1)
	case "seeding":
		atomic.AddInt64(&m.TorrentsSeeding, 1)
	case "stopped":
		atomic.AddInt64(&m.TorrentsStopped, 1)
	case "error":
		atomic.AddInt64(&m.TorrentsError, 1)
	}
}

// AddBytesDownloaded adds to the downloaded bytes counter
func (m *Metrics) AddBytesDownloaded(bytes int64) {
	atomic.AddInt64(&m.BytesDownloaded, bytes)
}

// AddBytesUploaded adds to the uploaded bytes counter
func (m *Metrics) AddBytesUploaded(bytes int64) {
	atomic.AddInt64(&m.BytesUploaded, bytes)
}

// IncrementHTTPRequests increments the HTTP request counter
func (m *Metrics) IncrementHTTPRequests() {
	atomic.AddInt64(&m.HTTPRequests, 1)
}

// IncrementHTTPErrors increments the HTTP error counter
func (m *Metrics) IncrementHTTPErrors() {
	atomic.AddInt64(&m.HTTPErrors, 1)
}

// RecordHTTPDuration records an HTTP request duration
func (m *Metrics) RecordHTTPDuration(duration time.Duration) {
	m.HTTPRequestDuration.Record(duration)
}

// SetMemoryUsage updates the memory usage metric
func (m *Metrics) SetMemoryUsage(bytes int64) {
	atomic.StoreInt64(&m.MemoryUsage, bytes)
}

// SetGoroutineCount updates the goroutine count
func (m *Metrics) SetGoroutineCount(count int32) {
	atomic.StoreInt32(&m.GoroutineCount, count)
}

// Record records a duration measurement
func (d *DurationMetric) Record(duration time.Duration) {
	nanos := duration.Nanoseconds()

	d.mu.Lock()
	defer d.mu.Unlock()

	d.count++
	d.sum += nanos

	if nanos < d.min {
		d.min = nanos
	}
	if nanos > d.max {
		d.max = nanos
	}
}

// Stats returns duration statistics
func (d *DurationMetric) Stats() (count int64, avg, min, max time.Duration) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	count = d.count
	if count > 0 {
		avg = time.Duration(d.sum / count)
		minDuration := time.Duration(d.min)
		min = minDuration
		maxDuration := time.Duration(d.max)
		max = maxDuration
	}
	return
}

// Snapshot returns a snapshot of all metrics
func (m *Metrics) Snapshot() map[string]interface{} {
	httpCount, httpAvg, httpMin, httpMax := m.HTTPRequestDuration.Stats()

	return map[string]interface{}{
		"torrents": map[string]int64{
			"total":       atomic.LoadInt64(&m.TorrentsTotal),
			"downloading": atomic.LoadInt64(&m.TorrentsDownloading),
			"seeding":     atomic.LoadInt64(&m.TorrentsSeeding),
			"stopped":     atomic.LoadInt64(&m.TorrentsStopped),
			"error":       atomic.LoadInt64(&m.TorrentsError),
		},
		"transfer": map[string]int64{
			"downloaded": atomic.LoadInt64(&m.BytesDownloaded),
			"uploaded":   atomic.LoadInt64(&m.BytesUploaded),
		},
		"http": map[string]interface{}{
			"requests": atomic.LoadInt64(&m.HTTPRequests),
			"errors":   atomic.LoadInt64(&m.HTTPErrors),
			"duration": map[string]interface{}{
				"count": httpCount,
				"avg":   httpAvg.String(),
				"min":   httpMin.String(),
				"max":   httpMax.String(),
			},
		},
		"system": map[string]interface{}{
			"memory_bytes":   atomic.LoadInt64(&m.MemoryUsage),
			"goroutines":     atomic.LoadInt32(&m.GoroutineCount),
			"uptime_seconds": time.Since(m.StartTime).Seconds(),
		},
	}
}
