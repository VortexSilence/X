package encoding

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeCycle(t *testing.T) {
	original := []byte("this is a secret message")
	key := []byte("thisis32bitlongpassphraseimusing")

	// Encode
	encoded := NewEncodeZ(original, key).
		EnAES().
		EnCha().
		Compress().
		Build()

	if encoded == nil {
		t.Fatal("Encoding failed: result is nil")
	}

	// Decode
	decoded := NewDZ(encoded, key).
		Decompress().
		DecodeCha().
		DeAES().
		Build()

	if decoded == nil {
		t.Fatal("Decoding failed: result is nil")
	}

	// Check equality
	if !bytes.Equal(decoded, original) {
		t.Errorf("Decoded message does not match original.\nExpected: %s\nGot: %s", original, decoded)
	}
}

func BenchmarkEncodeDecodeCycle(b *testing.B) {
	original := []byte("this is a secret message")
	key := []byte("thisis32bitlongpassphraseimusing")

	for i := 0; i < b.N; i++ {
		// Encode
		encoded := NewEncodeZ(original, key).
			EnAES().
			EnCha().
			Compress().
			Build()

		if encoded == nil {
			b.Fatal("Encoding failed: result is nil")
		}

		// Decode
		decoded := NewDZ(encoded, key).
			Decompress().
			DecodeCha().
			DeAES().
			Build()

		if decoded == nil {
			b.Fatal("Decoding failed: result is nil")
		}
	}
}
