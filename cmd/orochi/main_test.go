package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	t.Run("バージョン情報を表示できる", func(t *testing.T) {
		// NOTE: アプリケーションの基本的な動作確認.
		// 実際のテストは統合テストで行う.
		if version == "" {
			t.Error("version should not be empty")
		}
	})
}

func TestShowDisclaimer(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	showDisclaimer()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected text
	expectedPhrases := []string{
		"OROCHI - DISCLAIMER",
		"BitTorrent",
		"copyrighted material",
		"legal purposes",
		"developers are not responsible",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Disclaimer missing expected phrase: %q", phrase)
		}
	}

	// Verify the disclaimer has proper formatting
	if !strings.Contains(output, "================================================================================") {
		t.Error("Disclaimer missing separator lines")
	}
}

func TestVersionInfo(t *testing.T) {
	// Test that version variables are properly initialized
	tests := []struct {
		name     string
		variable string
		value    string
	}{
		{"version", "version", version},
		{"commit", "commit", commit},
		{"date", "date", date},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s should not be empty", tt.variable)
			}
		})
	}
}
