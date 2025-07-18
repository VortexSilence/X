package shadowsocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"log"
)

func AEAD(p string) cipher.AEAD {
	key := sha256.Sum256([]byte(p))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		log.Fatal(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}
	return aead
}
