package padding

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
)

const size = 512

func Encode(data []byte) ([]byte, error) {
	dataLen := len(data)
	if dataLen > size-2 {
		return nil, fmt.Errorf("data too long: max %d bytes", size-2)
	}
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, uint16(dataLen)); err != nil {
		return nil, err
	}
	buf.Write(data)
	paddingSize := size - 2 - dataLen
	padding := make([]byte, paddingSize)
	if _, err := rand.Read(padding); err != nil {
		return nil, err
	}
	buf.Write(padding)
	return buf.Bytes(), nil
}

func Decode(encoded []byte) ([]byte, error) {
	if len(encoded) != size {
		return nil, errors.New("invalid message size")
	}
	dataLen := binary.BigEndian.Uint16(encoded[:2])
	if int(dataLen) > size-2 {
		return nil, errors.New("length field too large")
	}
	data := encoded[2 : 2+dataLen]
	return data, nil
}
