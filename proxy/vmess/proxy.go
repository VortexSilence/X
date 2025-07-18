package vmess

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"sync"
	"time"

)

// Encoder holds IV and key
type Encoder struct {
	IV  []byte // 16 bytes
	Key []byte // 16 bytes
}

func NewEncoder() (*Encoder, error) {
	iv := make([]byte, 16)
	key := make([]byte, 16)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	_, err = rand.Read(key)
	if err != nil {
		return nil, err
	}
	return &Encoder{
		IV:  iv,
		Key: key,
	}, nil
}

// VmessWriter represents the write side of the VMess connection
type VmessWriter struct {
	mu         sync.Mutex
	conn       net.Conn
	uuid       []byte // 16 bytes
	encoder    *Encoder
	handshaked bool
	aead       bool
}

func NewVmessWriter(conn net.Conn, uuid []byte, aead bool) (*VmessWriter, error) {
	enc, err := NewEncoder()
	if err != nil {
		return nil, err
	}
	return &VmessWriter{
		conn:    conn,
		uuid:    uuid,
		encoder: enc,
		aead:    aead,
	}, nil
}

// MD5 helper
func md5Sum(data ...[]byte) []byte {
	h := md5.New()
	for _, b := range data {
		h.Write(b)
	}
	return h.Sum(nil)
}

// FNV1a 32-bit
func fnv1aHash32(data []byte) uint32 {
	const prime32 = 16777619
	var hash uint32 = 2166136261

	for _, b := range data {
		hash ^= uint32(b)
		hash *= prime32
	}
	return hash
}

// AES-128-CFB encryptor
type aesCFBEncryptor struct {
	stream cipher.Stream
}

func newAesCFBEncryptor(key, iv []byte) (*aesCFBEncryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	return &aesCFBEncryptor{stream}, nil
}

func (a *aesCFBEncryptor) Encrypt(data []byte) {
	a.stream.XORKeyStream(data, data)
}

func (w *VmessWriter) handshake(ctx context.Context, host string, port uint16) error {
	// timestamp in seconds big endian
	now := uint64(time.Now().Unix())
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, now)

	if !w.aead {
		mac := hmac.New(md5.New, w.uuid)
		mac.Write(timeBytes)
		auth := mac.Sum(nil) // 16 bytes
		// send auth
		if _, err := w.conn.Write(auth); err != nil {
			return err
		}
	}

	// build cmd buffer
	cmd := []byte{0x1} // version=1
	cmd = append(cmd, w.encoder.IV...)
	cmd = append(cmd, w.encoder.Key...)
	if !w.aead {
		cmd = append(cmd,
			0x00, // Response Authentication Value
			0x01, // Option S (Standard format)
			0x00, // Encryption Method AES-128-CFB
			0x00, // reserved
			0x01, // Command TCP
		)
	} else {
		cmd = append(cmd,
			0x00, // Response Authentication Value
			0x01, // Option S
			0x05, // Encryption Method None
			0x00,
			0x01,
		)
	}
	// port (2 bytes BE)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	cmd = append(cmd, portBytes...)
	// address type domain = 0x02
	cmd = append(cmd, 0x02)
	// address length + address bytes
	cmd = append(cmd, byte(len(host)))
	cmd = append(cmd, []byte(host)...)
	// checksum fnv1a 32
	cs := fnv1aHash32(cmd)
	csBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(csBytes, cs)
	cmd = append(cmd, csBytes...)

	if !w.aead {
		// encryption key and iv for AES-128-CFB
		key := md5Sum(w.uuid, []byte("c48619fe-8f02-49e0-b9e9-edf763e17e21"))
		iv := md5Sum(timeBytes, timeBytes, timeBytes, timeBytes)

		enc, err := newAesCFBEncryptor(key, iv)
		if err != nil {
			return err
		}
		enc.Encrypt(cmd)
	} else {
		// AEAD encryption - complex, for now not implemented
		return errors.New("AEAD handshake not implemented")
	}

	// write cmd buffer
	_, err := w.conn.Write(cmd)
	return err
}

// Write data through VMess connection
func (w *VmessWriter) Write(ctx context.Context, data []byte, host string, port uint16) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.handshaked {
		if err := w.handshake(ctx, host, port); err != nil {
			return err
		}
		w.handshaked = true
	}

	length := uint16(len(data))
	var vmessBuf bytes.Buffer
	if !w.aead {
		cs := fnv1aHash32(data)
		binary.Write(&vmessBuf, binary.BigEndian, length+4) // total length
		binary.Write(&vmessBuf, binary.BigEndian, cs)       // checksum
		vmessBuf.Write(data)

		key := w.encoder.Key
		iv := w.encoder.IV
		enc, err := newAesCFBEncryptor(key, iv)
		if err != nil {
			return err
		}
		buf := vmessBuf.Bytes()
		enc.Encrypt(buf)
		_, err = w.conn.Write(buf)
		return err
	} else {
		// AEAD mode - not implemented here
		return errors.New("AEAD mode not implemented")
	}
}
