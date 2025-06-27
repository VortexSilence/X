package z

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	NonceSize = chacha20poly1305.NonceSize
	KeySize   = chacha20poly1305.KeySize
	HMACSize  = 32
)

// Encrypt encrypts, obfuscates, and signs the message
func Encrypt(data, key, hmacKey []byte, xorSeed byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, errors.New("invalid key length")
	}
	if len(hmacKey) != 32 {
		return nil, errors.New("invalid hmac key length")
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, NonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// Encrypt
	cipher := aead.Seal(nil, nonce, data, nil)

	// Obfuscate
	for i := range cipher {
		cipher[i] ^= xorSeed + byte(i%13)
	}

	// Create HMAC
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipher)
	tag := mac.Sum(nil)

	// Final output: [nonce][obfuscated cipher][hmac]
	result := append(nonce, cipher...)
	result = append(result, tag...)
	return result, nil
}

// Decrypt verifies, de-obfuscates, and decrypts the message
func Decrypt(input, key, hmacKey []byte, xorSeed byte) ([]byte, error) {
	if len(input) < NonceSize+HMACSize {
		return nil, errors.New("input too short")
	}
	if len(key) != KeySize || len(hmacKey) != 32 {
		return nil, errors.New("invalid key or hmac key")
	}

	nonce := input[:NonceSize]
	cipher := input[NonceSize : len(input)-HMACSize]
	tag := input[len(input)-HMACSize:]

	// Verify HMAC
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write(cipher)
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, tag) {
		return nil, errors.New("hmac mismatch")
	}

	// De-obfuscate
	for i := range cipher {
		cipher[i] ^= xorSeed + byte(i%13)
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	plain, err := aead.Open(nil, nonce, cipher, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
