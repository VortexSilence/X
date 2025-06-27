package zdns

import (
	"encoding/base32"
	"fmt"
	"math/rand"

	"github.com/miekg/dns"
	"golang.org/x/crypto/chacha20poly1305"
)

type ZDNS struct {
	key []byte
}

func New(key []byte) (*ZDNS, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size")
	}
	return &ZDNS{key: key}, nil
}

// Encrypt and wrap inside fake DNS Query
func (z *ZDNS) EncodeDNSPacket(message []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(z.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)
	rand.Read(nonce)

	enc := aead.Seal(nil, nonce, message, nil)

	full := append(nonce, enc...) // prepend nonce

	// Base32 encode for QNAME
	b32 := base32.StdEncoding.WithPadding(base32.NoPadding)
	encoded := b32.EncodeToString(full)

	// Split into fake QNAME like abc.def.ghi.example.com
	qname := ""
	for i := 0; i < len(encoded); i += 10 {
		end := i + 10
		if end > len(encoded) {
			end = len(encoded)
		}
		qname += encoded[i:end] + "."
	}
	qname += "example.com." // fake domain

	// Create DNS message
	m := new(dns.Msg)
	m.SetQuestion(qname, dns.TypeA)

	return m.Pack()
}

// Decode fake DNS to original message
func (z *ZDNS) DecodeDNSPacket(packet []byte) ([]byte, error) {
	var m dns.Msg
	if err := m.Unpack(packet); err != nil {
		return nil, err
	}
	if len(m.Question) == 0 {
		return nil, fmt.Errorf("no question section")
	}

	// Reconstruct encoded string from QNAME
	qname := m.Question[0].Name
	trimmed := qname[:len(qname)-len(".example.com.")] // remove suffix
	encoded := ""
	for _, part := range dns.SplitDomainName(trimmed) {
		encoded += part
	}

	b32 := base32.StdEncoding.WithPadding(base32.NoPadding)
	cipherData, err := b32.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	if len(cipherData) < chacha20poly1305.NonceSize {
		return nil, fmt.Errorf("invalid encrypted message")
	}

	nonce := cipherData[:chacha20poly1305.NonceSize]
	ciphertext := cipherData[chacha20poly1305.NonceSize:]

	aead, err := chacha20poly1305.New(z.key)
	if err != nil {
		return nil, err
	}

	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
