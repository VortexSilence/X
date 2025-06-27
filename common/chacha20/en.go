package chacha20

import (
	"crypto/rand"

	"golang.org/x/crypto/chacha20poly1305"
)

func Encrypt(payload []byte, key []byte) ([]byte, error) {
	nonce := make([]byte, 12)
	_, _ = rand.Read(nonce)

	aead, _ := chacha20poly1305.New(key)
	ciphertext := aead.Seal(nil, nonce, payload, nil)

	return append(nonce, ciphertext...), nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	nonce := data[:12]
	ciphertext := data[12:]

	aead, _ := chacha20poly1305.New(key)
	return aead.Open(nil, nonce, ciphertext, nil)
}
