package padding

import (
	"testing"
)

func TestEncodeDecodeMessage(t *testing.T) {
	original := []byte("پیام تستی برای بررسی Encode و Decode")

	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}

	if len(encoded) != size {
		t.Errorf("Expected encoded size %d, got %d", size, len(encoded))
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("DecodeMessage failed: %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("Decoded message doesn't match original.\nExpected: %s\nGot: %s", original, decoded)
	}
}
