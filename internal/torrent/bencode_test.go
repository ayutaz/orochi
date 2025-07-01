package torrent

import (
	"testing"

	"github.com/zeebo/bencode"
)

func TestBencodeStructures(t *testing.T) {
	// Test that our structures can be properly encoded/decoded
	t.Run("ベンコード構造体が正しくエンコード/デコードできる", func(t *testing.T) {
		original := bencodeTorrent{
			Announce: "http://example.com:8000",
			Info: bencodeInfo{
				Name:        "test.txt",
				PieceLength: 16384,
				Length:      1024,
				Pieces:      "01234567890123456789",
			},
		}

		// Encode
		data, err := bencode.EncodeBytes(original)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}

		// Decode
		var decoded bencodeTorrent
		if err := bencode.DecodeBytes(data, &decoded); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		// Verify
		if decoded.Announce != original.Announce {
			t.Errorf("announce mismatch: got %s, want %s", decoded.Announce, original.Announce)
		}
		if decoded.Info.Name != original.Info.Name {
			t.Errorf("name mismatch: got %s, want %s", decoded.Info.Name, original.Info.Name)
		}
		if decoded.Info.Length != original.Info.Length {
			t.Errorf("length mismatch: got %d, want %d", decoded.Info.Length, original.Info.Length)
		}
		if decoded.Info.PieceLength != original.Info.PieceLength {
			t.Errorf("piece length mismatch: got %d, want %d", decoded.Info.PieceLength, original.Info.PieceLength)
		}
		if decoded.Info.Pieces != original.Info.Pieces {
			t.Errorf("pieces mismatch: got %s, want %s", decoded.Info.Pieces, original.Info.Pieces)
		}
	})
}
