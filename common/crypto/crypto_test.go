package crypto

import (
	"strings"
	"testing"
)

func TestEncryptDecryptAES256CBC(t *testing.T) {
	tests := []struct {
		name      string
		key       []byte
		plaintext []byte
		wantErr   bool
	}{
		{
			name:      "Valid AES-256 encryption/decryption",
			key:       []byte("thisis32bitlongpassphraseimusing"), // 32 bytes
			plaintext: []byte("سلام دنیا!"),
			wantErr:   false,
		},
		{
			name:      "Empty plaintext",
			key:       []byte("thisis32bitlongpassphraseimusing"),
			plaintext: []byte(""),
			wantErr:   false,
		},
		{
			name:      "Invalid key size (too short)",
			key:       []byte("shortkey"),
			plaintext: []byte("test"),
			wantErr:   true,
		},
		{
			name:      "Invalid key size (too long)",
			key:       []byte("thiskeyistoolongandhasmorethan32bytes!!!"),
			plaintext: []byte("test"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptAES256CBC([]byte(tt.plaintext), tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Encrypt error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return // اگر انتظار خطا داشتیم، اینجا تموم می‌کنیم
			}

			decrypted, err := DecryptAES256CBC(encrypted, tt.key)
			if err != nil {
				t.Fatalf("Decrypt error = %v", err)
			}
			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("Decrypted text = %s, want = %s", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAES256CBC(t *testing.T) {
	key := []byte("thisis32bitlongpassphraseimusing") // 32 bytes
	plaintext := "سلام دنیا!"

	ciphertext, err := EncryptAES256CBC([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptAES256CBC(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match. Got %s, expected %s", decrypted, plaintext)
	}
}

func BenchmarkEncryptAES256CBC(b *testing.B) {
	key := []byte("thisis32bitlongpassphraseimusing")          // 32 bytes
	plaintext := []byte(strings.Repeat("Hello, AES-256!", 20)) // ~280 bytes

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptAES256CBC(plaintext, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkDecryptAES256CBC(b *testing.B) {
	key := []byte("thisis32bitlongpassphraseimusing")
	plaintext := []byte(strings.Repeat("Hello, AES-256!", 20))

	// Encrypt once for decryption benchmark
	encrypted, err := EncryptAES256CBC(plaintext, key)
	if err != nil {
		b.Fatalf("Encryption for benchmark setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptAES256CBC(encrypted, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}
