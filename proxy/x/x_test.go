package x

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeCycle(t *testing.T) {
	original := []byte("this is a secret message")
	key := []byte("0123456789abcdef0123456789abcdef") // طول 32 بایت برای chacha20

	// Encode
	encoded := NewEncodeX(original, key).
		EnCha().
		Compress().
		Build()

	if encoded == nil {
		t.Fatal("Encoding failed: result is nil")
	}

	// Decode
	decoded := NewDX(encoded, key).
		Decompress().
		DesktopCha().
		Build()

	if decoded == nil {
		t.Fatal("Decoding failed: result is nil")
	}

	// Check equality
	if !bytes.Equal(decoded, original) {
		t.Errorf("Decoded message does not match original.\nExpected: %s\nGot: %s", original, decoded)
	}
}

var (
	sampleMessage = []byte("this is a secret message that will be repeatedly encoded and decoded under heavy load")
	sampleKey     = []byte("0123456789abcdef0123456789abcdef") // 32 bytes for ChaCha20
)

func BenchmarkEncodeDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Encode
		encoded := NewEncodeX(sampleMessage, sampleKey).
			EnCha().
			Compress().
			Build()

		// Decode
		decoded := NewDX(encoded, sampleKey).
			Decompress().
			DesktopCha().
			Build()

		// Optional: Validate result (can be removed to improve speed)
		if !bytes.Equal(decoded, sampleMessage) {
			b.Fatal("Mismatch between original and decoded message")
		}
	}
}
