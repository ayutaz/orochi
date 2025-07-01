package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	t.Run("デフォルト設定を読み込める", func(t *testing.T) {
		// Red phase: このテストは現在失敗する
		config := LoadDefault()

		if config == nil {
			t.Fatal("config should not be nil")
		}

		if config.Port != 8080 {
			t.Errorf("expected port 8080, got %d", config.Port)
		}

		if config.DownloadDir == "" {
			t.Error("download directory should not be empty")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "有効な設定",
			config: &Config{
				Port:        8080,
				DownloadDir: "./downloads",
				MaxTorrents: 5,
				MaxPeers:    200,
			},
			wantErr: false,
		},
		{
			name: "無効なポート番号（0）",
			config: &Config{
				Port:        0,
				DownloadDir: "./downloads",
				MaxTorrents: 5,
				MaxPeers:    200,
			},
			wantErr: true,
		},
		{
			name: "無効なポート番号（65536以上）",
			config: &Config{
				Port:        65536,
				DownloadDir: "./downloads",
				MaxTorrents: 5,
				MaxPeers:    200,
			},
			wantErr: true,
		},
		{
			name: "空のダウンロードディレクトリ",
			config: &Config{
				Port:        8080,
				DownloadDir: "",
				MaxTorrents: 5,
				MaxPeers:    200,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetAbsoluteDownloadDir(t *testing.T) {
	t.Run("相対パスを絶対パスに変換する", func(t *testing.T) {
		config := &Config{
			DownloadDir: "./downloads",
		}

		absPath := config.GetAbsoluteDownloadDir()

		if !filepath.IsAbs(absPath) {
			t.Errorf("expected absolute path, got %s", absPath)
		}
	})

	t.Run("既に絶対パスの場合はそのまま返す", func(t *testing.T) {
		// Windowsでは異なるパス形式を使用
		var expectedPath string
		if runtime.GOOS == "windows" {
			expectedPath = `C:\tmp\downloads`
		} else {
			expectedPath = "/tmp/downloads"
		}
		
		config := &Config{
			DownloadDir: expectedPath,
		}

		absPath := config.GetAbsoluteDownloadDir()

		if absPath != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, absPath)
		}
	})
}
