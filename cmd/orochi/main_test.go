package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	t.Run("バージョン情報を表示できる", func(t *testing.T) {
		// アプリケーションの基本的な動作確認
		// 実際のテストは統合テストで行う
		if version == "" {
			t.Error("version should not be empty")
		}
	})
}