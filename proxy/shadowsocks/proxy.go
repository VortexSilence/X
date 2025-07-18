package shadowsocks

import (
	"crypto/cipher"
	"net"
)

func Encode(dst net.Conn, src net.Conn, aead cipher.AEAD, baseNonce []byte) {
	buf := make([]byte, 4096)
	nonce := make([]byte, len(baseNonce))
	copy(nonce, baseNonce)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		enc := aead.Seal(nil, nonce, buf[:n], nil)
		dst.Write(enc)
		increment(nonce)
	}
}

func Decode(dst net.Conn, src net.Conn, aead cipher.AEAD, baseNonce []byte) {
	nonce := make([]byte, len(baseNonce))
	copy(nonce, baseNonce)
	tag := aead.Overhead()
	buf := make([]byte, 4096+tag)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		dec, err := aead.Open(nil, nonce, buf[:n], nil)
		if err != nil {
			return
		}
		dst.Write(dec)
		increment(nonce)
	}
}

func increment(nonce []byte) {
	for i := len(nonce) - 1; i >= 0; i-- {
		nonce[i]++
		if nonce[i] != 0 {
			break
		}
	}
}
