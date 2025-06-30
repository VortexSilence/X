package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// pad adds PKCS#7 padding
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// unpad removes PKCS#7 padding
func unpad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, fmt.Errorf("unpad error: input too short")
	}
	unpadding := int(src[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("unpad error: invalid padding")
	}
	return src[:(length - unpadding)], nil
}

// EncryptAES256CBC encrypts plaintext using AES-256-CBC and returns base64-encoded ciphertext.
func EncryptAES256CBC(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pad(plaintext, aes.BlockSize)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
	// return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES256CBC decrypts base64-encoded ciphertext using AES-256-CBC.
func DecryptAES256CBC(ciphertext, key []byte) ([]byte, error) {
	// ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	// if err != nil {
	// 	return nil, err
	// }

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	plaintext, err := unpad(ciphertext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
