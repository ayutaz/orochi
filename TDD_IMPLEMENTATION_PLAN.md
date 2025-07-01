# TDDによるSimpleTorrent MVP実装計画

## TDDの基本原則（t-wadaスタイル）

### Red-Green-Refactorサイクル
1. **Red**: 失敗するテストを書く
2. **Green**: テストを通す最小限のコードを書く
3. **Refactor**: コードを改善する（テストは通ったまま）

### TDDの黄金律
- テストなしにプロダクションコードを書かない
- 失敗する単体テストを、成功させる以上のプロダクションコードを書かない
- 一度に複数のテストを書かない

## 実装順序（Outside-In TDD）

### フェーズ1: プロジェクトセットアップとCI/CD

#### 1.1 プロジェクト初期化
```bash
# 1. Goプロジェクトの初期化
mkdir simple-torrent && cd simple-torrent
go mod init github.com/yourusername/simple-torrent

# 2. 基本ディレクトリ構造の作成
mkdir -p cmd/simple-torrent
mkdir -p internal/{torrent,web,config}
mkdir -p web/static/{css,js}
mkdir -p web/templates
mkdir -p test/integration
```

#### 1.2 最初のテスト（プロジェクトの動作確認）
```go
// main_test.go
package main

import "testing"

func TestMain(t *testing.T) {
    // アプリケーションが起動することを確認
    t.Run("アプリケーションが正常に起動する", func(t *testing.T) {
        // ここから始める
    })
}
```

### フェーズ2: 設定管理（TDD）

#### 2.1 設定読み込みのテスト
```go
// internal/config/config_test.go
func TestLoadConfig(t *testing.T) {
    t.Run("デフォルト設定を読み込める", func(t *testing.T) {
        // Red: 失敗するテストを書く
        config := LoadDefault()
        assert.NotNil(t, config)
        assert.Equal(t, 8080, config.Port)
    })
}
```

#### 2.2 実装
```go
// internal/config/config.go
type Config struct {
    Port int
    DownloadDir string
    // 最小限の設定のみ
}

func LoadDefault() *Config {
    return &Config{
        Port: 8080,
        DownloadDir: "./downloads",
    }
}
```

### フェーズ3: Webサーバー（TDD）

#### 3.1 HTTPサーバーのテスト
```go
// internal/web/server_test.go
func TestServer(t *testing.T) {
    t.Run("サーバーが起動してヘルスチェックに応答する", func(t *testing.T) {
        // Red
        server := NewServer(":8080")
        req := httptest.NewRequest("GET", "/health", nil)
        w := httptest.NewRecorder()
        
        server.ServeHTTP(w, req)
        
        assert.Equal(t, 200, w.Code)
        assert.Equal(t, "OK", w.Body.String())
    })
}
```

### フェーズ4: トレント機能（コア機能のTDD）

#### 4.1 トレントファイルのパース
```go
// internal/torrent/parser_test.go
func TestParseTorrentFile(t *testing.T) {
    t.Run("有効なtorrentファイルをパースできる", func(t *testing.T) {
        // Red: テストデータを用意
        torrentData := []byte("d8:announce...")
        
        info, err := ParseTorrent(torrentData)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, info.Name)
        assert.NotEmpty(t, info.InfoHash)
    })
}
```

#### 4.2 マグネットリンクのパース
```go
func TestParseMagnetLink(t *testing.T) {
    t.Run("有効なマグネットリンクをパースできる", func(t *testing.T) {
        // Red
        magnetLink := "magnet:?xt=urn:btih:..."
        
        info, err := ParseMagnet(magnetLink)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, info.InfoHash)
    })
}
```

### フェーズ5: トレントマネージャー（統合）

#### 5.1 トレントの追加
```go
// internal/torrent/manager_test.go
func TestTorrentManager(t *testing.T) {
    t.Run("トレントを追加できる", func(t *testing.T) {
        // Red
        manager := NewManager()
        torrentData := []byte("...")
        
        id, err := manager.AddTorrent(torrentData)
        
        assert.NoError(t, err)
        assert.NotEmpty(t, id)
        assert.Equal(t, 1, manager.Count())
    })
}
```

### フェーズ6: REST API（E2Eに近いテスト）

#### 6.1 APIエンドポイントのテスト
```go
// internal/web/api_test.go
func TestAPI(t *testing.T) {
    t.Run("POST /api/torrents でトレントを追加できる", func(t *testing.T) {
        // Red
        server := setupTestServer()
        torrentFile := createTestTorrent()
        
        resp := postTorrent(server, torrentFile)
        
        assert.Equal(t, 201, resp.StatusCode)
        var result map[string]string
        json.NewDecoder(resp.Body).Decode(&result)
        assert.NotEmpty(t, result["id"])
    })
}
```

## テスト戦略

### 1. 単体テストのガイドライン
- **F.I.R.S.T原則**
  - Fast: 高速に実行
  - Independent: 独立して実行可能
  - Repeatable: 何度でも同じ結果
  - Self-validating: 自動で検証
  - Timely: タイムリーに書く

### 2. テストダブルの使用
```go
// モックの例
type MockTorrentClient struct {
    mock.Mock
}

func (m *MockTorrentClient) Download(infoHash string) error {
    args := m.Called(infoHash)
    return args.Error(0)
}
```

### 3. テーブル駆動テスト
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"有効なマグネットリンク", "magnet:?xt=...", false},
        {"無効なマグネットリンク", "http://...", true},
        {"空の入力", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateMagnet(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## MVP実装タイムライン

### Week 1-2: 基礎
- [ ] プロジェクトセットアップ
- [ ] CI/CDパイプライン構築
- [ ] 設定管理のTDD実装
- [ ] 基本的なWebサーバー

### Week 3-4: コア機能
- [ ] トレントファイルパーサー
- [ ] マグネットリンクパーサー
- [ ] 基本的なトレントマネージャー

### Week 5-6: API実装
- [ ] REST APIエンドポイント
- [ ] WebSocket通信
- [ ] エラーハンドリング

### Week 7-8: UI実装
- [ ] 基本的なWeb UI
- [ ] リアルタイム更新
- [ ] レスポンシブデザイン

### Week 9-10: 統合とセキュリティ
- [ ] anacrolix/torrentライブラリ統合
- [ ] VPNバインディング
- [ ] IPフィルタリング

### Week 11-12: 仕上げ
- [ ] E2Eテスト
- [ ] パフォーマンステスト
- [ ] ドキュメント作成
- [ ] リリース準備

## テスト実行とカバレッジ

### Makefile
```makefile
.PHONY: test coverage lint

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

test-watch:
	gotestsum --watch
```

## TDD実装のベストプラクティス

### 1. コミットメッセージ
```
feat: トレントファイルパーサーを追加

- torrentファイルのベンコードをデコード
- 必須フィールドの検証を実装
- 不正なファイルに対するエラーハンドリング

テストケース:
- 正常なtorrentファイル
- 破損したtorrentファイル
- 必須フィールドが欠けているファイル
```

### 2. テストの命名規則
```go
// 良い例
func TestTorrentManager_AddTorrent_ValidFile_ReturnsID(t *testing.T)
func TestTorrentManager_AddTorrent_InvalidFile_ReturnsError(t *testing.T)

// 日本語も可（t-wadaスタイル）
func Test_トレントマネージャー_正常なファイルを追加できる(t *testing.T)
```

### 3. AAA パターン
```go
func TestSomething(t *testing.T) {
    // Arrange（準備）
    manager := NewManager()
    testFile := createTestFile()
    
    // Act（実行）
    result, err := manager.Process(testFile)
    
    // Assert（検証）
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

---

この計画に従ってTDDでMVPを実装していきます。各フェーズで必ずテストファーストで進め、シンプルで保守性の高いコードを目指します。