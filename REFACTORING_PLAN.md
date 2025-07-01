# Orochi リファクタリング計画

## 🎯 リファクタリングの目的

1. **テスタビリティの向上**: インターフェースの導入によるモックの容易化
2. **保守性の向上**: 責任の分離と疎結合化
3. **拡張性の向上**: 新機能追加を容易にする設計
4. **エラーハンドリングの改善**: 一貫性のあるエラー処理
5. **パフォーマンスの最適化**: 並行処理の改善

## 📋 リファクタリング項目（優先度順）

### 1. インターフェースの導入 🔴 最優先

#### 現状の問題点
- 具象型への直接依存によりテストが困難
- モックが作成できない
- 依存性注入ができない

#### 改善案

```go
// internal/torrent/interfaces.go
package torrent

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

type Parser interface {
    ParseTorrentFile(data []byte) (*TorrentInfo, error)
    ParseMagnetLink(magnetLink string) (*TorrentInfo, error)
}

type Storage interface {
    Save(torrent *Torrent) error
    Load(id string) (*Torrent, error)
    Delete(id string) error
    List() ([]*Torrent, error)
}
```

### 2. エラーハンドリングの統一化 🔴 最優先

#### 現状の問題点
- 文字列エラーの多用
- エラーの種類が判別できない
- 一貫性のないエラーメッセージ

#### 改善案

```go
// internal/errors/errors.go
package errors

import "fmt"

type ErrorCode string

const (
    ErrCodeNotFound      ErrorCode = "NOT_FOUND"
    ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
    ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
    ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
)

type AppError struct {
    Code    ErrorCode
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructors
func NotFound(message string) *AppError {
    return &AppError{Code: ErrCodeNotFound, Message: message}
}

func InvalidInput(message string) *AppError {
    return &AppError{Code: ErrCodeInvalidInput, Message: message}
}
```

### 3. ロギングシステムの導入 🟡 高優先

#### 現状の問題点
- 標準の`log`パッケージのみ使用
- 構造化ログなし
- ログレベルの制御なし

#### 改善案

```go
// internal/logger/logger.go
package logger

import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

type Field struct {
    Key   string
    Value interface{}
}

func New(level string) Logger {
    // zerologの実装
}
```

### 4. HTTPルーターの改善 🟡 高優先

#### 現状の問題点
- 標準の`http.ServeMux`使用で機能が限定的
- ミドルウェアサポートなし
- パスパラメータの手動パース

#### 改善案

```go
// internal/web/router.go
package web

import "github.com/go-chi/chi/v5"

func (s *Server) setupRoutes() {
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(s.requireTorrentManager)
    
    // Routes
    r.Route("/api", func(r chi.Router) {
        r.Route("/torrents", func(r chi.Router) {
            r.Get("/", s.handleListTorrents)
            r.Post("/", s.handleAddTorrent)
            r.Post("/magnet", s.handleAddMagnet)
            
            r.Route("/{id}", func(r chi.Router) {
                r.Get("/", s.handleGetTorrent)
                r.Delete("/", s.handleDeleteTorrent)
                r.Post("/start", s.handleStartTorrent)
                r.Post("/stop", s.handleStopTorrent)
            })
        })
    })
}
```

### 5. 並行処理の最適化 🟡 高優先

#### 現状の問題点
- 単純な`sync.RWMutex`使用
- 読み取り操作でも書き込みロックを取得する場合がある

#### 改善案

```go
// internal/torrent/manager.go
func (m *Manager) GetTorrent(id string) (*Torrent, bool) {
    m.mu.RLock() // 読み取りロックを使用
    defer m.mu.RUnlock()
    
    torrent, exists := m.torrents[id]
    return torrent, exists
}

// sync.Mapの使用も検討
type Manager struct {
    torrents sync.Map // より効率的な並行アクセス
}
```

### 6. 設定管理の改善 🟢 中優先

#### 現状の問題点
- 設定ファイルのサポートなし
- 環境変数のサポートなし
- ハードコードされた値

#### 改善案

```go
// internal/config/loader.go
package config

import "github.com/spf13/viper"

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.orochi")
    
    // 環境変数のバインド
    viper.BindEnv("port", "OROCHI_PORT")
    viper.BindEnv("download_dir", "OROCHI_DOWNLOAD_DIR")
    
    // デフォルト値
    viper.SetDefault("port", 8080)
    viper.SetDefault("download_dir", "./downloads")
    
    if err := viper.ReadInConfig(); err != nil {
        // 設定ファイルがない場合はデフォルト値を使用
    }
    
    var cfg Config
    return &cfg, viper.Unmarshal(&cfg)
}
```

### 7. テストの改善 🟢 中優先

#### 現状の問題点
- 統合テストの不足
- モックの欠如
- ベンチマークテストなし

#### 改善案

```go
// internal/torrent/manager_test.go
func TestManagerConcurrency(t *testing.T) {
    manager := NewManager()
    
    // 並行アクセステスト
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            
            // 並行して追加・取得・削除
            id, _ := manager.AddMagnet(fmt.Sprintf("magnet:?xt=urn:btih:%040d", i))
            manager.GetTorrent(id)
            manager.RemoveTorrent(id)
        }(i)
    }
    wg.Wait()
}

// ベンチマーク
func BenchmarkManagerAdd(b *testing.B) {
    manager := NewManager()
    data := CreateTestTorrent()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        manager.AddTorrent(data)
    }
}
```

### 8. APIドキュメンテーション 🟢 中優先

#### 現状の問題点
- API仕様書なし
- エンドポイントの説明不足

#### 改善案

```go
// internal/web/docs.go
// swaggoを使用したドキュメント生成

// @Summary List all torrents
// @Description Get a list of all torrents
// @Tags torrents
// @Accept json
// @Produce json
// @Success 200 {array} torrent.Torrent
// @Router /api/torrents [get]
func (s *Server) handleListTorrents(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### 9. メトリクスの追加 🔵 低優先

#### 現状の問題点
- パフォーマンスモニタリングなし
- エラー率の可視化なし

#### 改善案

```go
// internal/metrics/metrics.go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    TorrentsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "orochi_torrents_total",
        Help: "Total number of torrents",
    })
    
    DownloadBytesTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "orochi_download_bytes_total",
        Help: "Total bytes downloaded",
    })
)
```

## 📊 リファクタリングのメトリクス

### 現在の状態
- テストカバレッジ: 60-80%
- 外部依存: 1個
- コード行数: 約2,000行

### 目標
- テストカバレッジ: 90%以上
- インターフェース化: 主要コンポーネントの100%
- エラーハンドリング: カスタムエラー型への完全移行
- ドキュメント: 全公開APIの文書化

## 🚀 実装順序

1. **Phase 1 (1週間)**
   - インターフェースの定義と実装
   - エラーハンドリングの統一化
   - 基本的なテストの追加

2. **Phase 2 (1週間)**
   - ロギングシステムの導入
   - HTTPルーターの改善
   - ミドルウェアの実装

3. **Phase 3 (3-4日)**
   - 並行処理の最適化
   - 設定管理の改善
   - 統合テストの追加

4. **Phase 4 (3-4日)**
   - APIドキュメンテーション
   - メトリクスの追加
   - パフォーマンステスト

---

最終更新: 2025-01-31