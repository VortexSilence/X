package vmess

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type VmessReader struct {
	conn net.Conn
	key  []byte // 16 bytes
	iv   []byte // 16 bytes
	aead bool
	dec  cipher.Stream
}

func NewVmessReader(conn net.Conn, key, iv []byte, aead bool) (*VmessReader, error) {
	if len(key) != 16 || len(iv) != 16 {
		return nil, errors.New("key and iv must be 16 bytes")
	}
	vr := &VmessReader{
		conn: conn,
		key:  key,
		iv:   iv,
		aead: aead,
	}

	if !aead {
		block, err := aes.NewCipher(md5Sum(key)) // md5 of key
		if err != nil {
			return nil, err
		}
		ivMd5 := md5Sum(iv) // md5 of iv
		vr.dec = cipher.NewCFBDecrypter(block, ivMd5)
	} else {
		// AEAD reading not implemented now
	}

	return vr, nil
}

// Read decrypts data from connection into buf, returns length of decrypted data
func (vr *VmessReader) Read(buf []byte) (int, error) {
	if vr.aead {
		return 0, errors.New("AEAD mode read not implemented")
	}

	// Step 1: read 4 byte header (encrypted)
	header := make([]byte, 4)
	if _, err := io.ReadFull(vr.conn, header); err != nil {
		return 0, err
	}
	vr.dec.XORKeyStream(header, header) // decrypt header but ignore content for now

	// Step 2: read 2 byte length (encrypted)
	lengthBytes := make([]byte, 2)
	if _, err := io.ReadFull(vr.conn, lengthBytes); err != nil {
		return 0, err
	}
	vr.dec.XORKeyStream(lengthBytes, lengthBytes)

	length := int(binary.BigEndian.Uint16(lengthBytes)) - 4 // remove checksum size

	if length > len(buf) {
		return 0, errors.New("buffer too small")
	}

	// Step 3: read 4 byte checksum (encrypted)
	checksum := make([]byte, 4)
	if _, err := io.ReadFull(vr.conn, checksum); err != nil {
		return 0, err
	}
	vr.dec.XORKeyStream(checksum, checksum)

	// Step 4: read actual data (encrypted)
	data := make([]byte, length)
	if _, err := io.ReadFull(vr.conn, data); err != nil {
		return 0, err
	}
	vr.dec.XORKeyStream(data, data)

	// Step 5: verify checksum
	cs := fnv1aHash32(data)
	expected := binary.BigEndian.Uint32(checksum)
	if cs != expected {
		return 0, errors.New("checksum mismatch")
	}

	copy(buf, data)
	return length, nil
}
