package z

import (
	"bytes"
	"testing"
)

var (
	testKey     = []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	testHMACKey = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 32 bytes
	testSeed    = byte(42)
	testMessage = []byte("This is a secret test message to be encrypted and obfuscated.")
)

// ✅ Unit Test
func TestEncryptDecrypt(t *testing.T) {
	cipher, err := Encrypt(testMessage, testKey, testHMACKey, testSeed)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	plain, err := Decrypt(cipher, testKey, testHMACKey, testSeed)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plain, testMessage) {
		t.Fatalf("Decrypted message does not match original.\nGot:  %s\nWant: %s", string(plain), string(testMessage))
	}
}

// ❌ HMAC Tamper Test
func TestTamperHMAC(t *testing.T) {
	cipher, err := Encrypt(testMessage, testKey, testHMACKey, testSeed)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with ciphertext
	cipher[len(cipher)-1] ^= 0xFF

	_, err = Decrypt(cipher, testKey, testHMACKey, testSeed)
	if err == nil {
		t.Fatal("Expected HMAC mismatch error, got nil")
	}
}

func BenchmarkEncrypt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Encrypt(testMessage, testKey, testHMACKey, testSeed)
		if err != nil {
			b.Fatal(err)
		}
	}
}
