package https

import (
	"encoding/binary"
	"math/rand"
)

type TLSCamouflage struct {
	SNIList      []string
	CipherSuites []uint16
}

func NewTLSCamouflage() *TLSCamouflage {
	return &TLSCamouflage{
		SNIList: []string{
			"cloudflare.com",
			"google.com",
			"apple.com",
			"microsoft.com",
		},
		CipherSuites: []uint16{
			0x1301, // TLS_AES_128_GCM_SHA256
			0x1302, // TLS_AES_256_GCM_SHA384
			0xC02B, // TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
			0xC02F, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		},
	}
}

func (t *TLSCamouflage) generateClientHello() []byte {
	buf := make([]byte, 0)
	// Record Layer (5 bytes)
	buf = append(buf, 0x16, 0x03, 0x01) // Handshake, TLS 1.0
	helloLength := make([]byte, 2)
	buf = append(buf, helloLength...) // Length (placeholder)

	// Handshake Header (4 bytes)
	buf = append(buf, 0x01) // ClientHello
	hsLength := make([]byte, 3)
	buf = append(buf, hsLength...) // Length (placeholder)

	// TLS Version (2 bytes)
	buf = append(buf, 0x03, 0x03) // TLS 1.2

	// Random (32 bytes)
	random := make([]byte, 32)
	rand.Read(random)
	buf = append(buf, random...)

	// Session ID (1 + 32 bytes)
	buf = append(buf, 0x20) // Length
	sessionID := make([]byte, 32)
	rand.Read(sessionID)
	buf = append(buf, sessionID...)

	// Cipher Suites (2 + n*2 bytes)
	cipherLen := len(t.CipherSuites) * 2
	buf = append(buf, byte(cipherLen>>8), byte(cipherLen))
	for _, suite := range t.CipherSuites {
		buf = append(buf, byte(suite>>8), byte(suite))
	}

	// Compression Methods (1 + 1 bytes)
	buf = append(buf, 0x01, 0x00) // Null compression

	// Extensions Length (2 bytes)
	extLength := make([]byte, 2)
	buf = append(buf, extLength...) // Placeholder

	// SNI Extension
	sni := t.SNIList[rand.Intn(len(t.SNIList))]
	buf = append(buf, 0x00, 0x00) // Extension Type (SNI)
	sniExt := make([]byte, 2)
	binary.BigEndian.PutUint16(sniExt, uint16(len(sni)+5))
	buf = append(buf, sniExt...)
	buf = append(buf, 0x00, byte(len(sni)+3), 0x00, byte(len(sni)))
	buf = append(buf, []byte(sni)...)

	// Update lengths
	// Extension Length
	extsLen := len(buf) - 2 - int(binary.BigEndian.Uint16(buf[len(buf)-2:]))
	binary.BigEndian.PutUint16(buf[len(buf)-extsLen-2:], uint16(extsLen))

	// Handshake Length
	hsLen := len(buf) - 5
	binary.BigEndian.PutUint16(buf[3:], uint16(hsLen))
	binary.BigEndian.PutUint32(buf[5:], uint32(hsLen))

	return buf
}

func (t *TLSCamouflage) Wrap(data []byte) []byte {
	clientHello := t.generateClientHello()
	wrapped := make([]byte, len(clientHello)+len(data))
	copy(wrapped, clientHello)
	copy(wrapped[len(clientHello):], data)
	return wrapped
}
