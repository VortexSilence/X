package codec

import (
	"testing"
)

func BenchmarkEncodeDecode(b *testing.B) {
	msg := []byte("salam ke raft toye hezar jaye mokhtalef, in test baraye benchmarck hast")

	for i := 0; i < b.N; i++ {
		// Encode
		packet := Encode(msg)
		// Decode
		decoded, err := Decode(packet)
		if err != nil {
			b.Fatalf("Decode failed: %v", err)
		}

		if string(decoded) != string(msg) {
			b.Fatalf("Decoded message mismatch")
		}
	}
}
