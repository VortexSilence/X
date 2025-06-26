package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
)

const (
	RealChunkType     = 0x1E
	TotalChunkTypes   = 40
	DecoyChunkCount   = 15
	MySQLProtocolByte = 0x0A
)

// ENCODE: ساخت پکت نهایی شبیه MySQL
func buildFinalMySQLLikePacket(data []byte) []byte {
	payload := new(bytes.Buffer)

	// --- MySQL-like handshake header ---
	payload.WriteByte(MySQLProtocolByte)
	payload.WriteString("5.7.31\x00") // version
	payload.Write(make([]byte, 20))   // filler
	payload.WriteByte(0x00)           // plugin data len
	payload.Write(make([]byte, 10))   // extra filler
	payload.WriteByte(0x00)           // capability flags

	// --- Chunks ---
	chunks := [][]byte{}

	// Step 1: double compression
	snappyCompressed := snappy.Encode(nil, data)
	lz4Compressed := compressLZ4(snappyCompressed)

	// Step 2: scrambling will happen inside buildChunk only
	mask := byte(rand.Intn(256))

	// Step 3: real chunk
	chunks = append(chunks, buildChunk(RealChunkType, lz4Compressed, mask))

	// Step 4: decoy chunks
	used := map[byte]bool{RealChunkType: true}
	for len(chunks) < DecoyChunkCount+1 {
		t := byte(rand.Intn(TotalChunkTypes) + 1)
		if used[t] {
			continue
		}
		used[t] = true
		fake := make([]byte, rand.Intn(10)+5)
		rand.Read(fake)
		fakeMask := byte(rand.Intn(256))
		chunks = append(chunks, buildChunk(t, fake, fakeMask))
	}

	// Step 5: shuffle chunks
	rand.Shuffle(len(chunks), func(i, j int) {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	})

	for _, c := range chunks {
		payload.Write(c)
	}

	// --- MySQL packet header ---
	final := new(bytes.Buffer)
	totalLen := payload.Len()
	final.WriteByte(byte(totalLen))
	final.WriteByte(byte(totalLen >> 8))
	final.WriteByte(byte(totalLen >> 16))
	final.WriteByte(0x00) // sequence ID
	final.Write(payload.Bytes())

	return final.Bytes()
}

// DECODE: استخراج پیام از پکت
func decodeFinalMySQLLikePacket(packet []byte) ([]byte, error) {
	if len(packet) < 4 {
		return nil, fmt.Errorf("packet too short")
	}

	payload := packet[4:]

	// Skip MySQL fake header
	offset := 1 + len("5.7.31") + 1 + 20 + 1 + 10 + 1
	if len(payload) <= offset {
		return nil, fmt.Errorf("payload too short")
	}

	reader := bytes.NewReader(payload[offset:])
	for reader.Len() > 0 {
		var t byte
		var flags byte
		var l uint16

		if err := binary.Read(reader, binary.BigEndian, &t); err != nil {
			break
		}
		if err := binary.Read(reader, binary.BigEndian, &flags); err != nil {
			break
		}
		if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
			break
		}

		data := make([]byte, l)
		n, err := reader.Read(data)
		if err != nil || n != int(l) {
			break
		}

		if t == RealChunkType {
			descrambled := xorBytes(data, flags)
			snappyData, err := decompressLZ4(descrambled)
			if err != nil {
				return nil, fmt.Errorf("lz4 decode failed: %v", err)
			}
			original, err := snappy.Decode(nil, snappyData)
			if err != nil {
				return nil, fmt.Errorf("snappy decode failed: %v", err)
			}
			return original, nil
		}
	}

	return nil, fmt.Errorf("no valid chunk found")
}

// ساخت یک chunk با scrambling و نوع مشخص
func buildChunk(t byte, data []byte, mask byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(t)
	buf.WriteByte(mask) // flags = scrambling mask
	binary.Write(buf, binary.BigEndian, uint16(len(data)))
	buf.Write(xorBytes(data, mask))
	return buf.Bytes()
}

// XOR scramble
func xorBytes(data []byte, mask byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ mask
	}
	return out
}

// LZ4 compress
func compressLZ4(data []byte) []byte {
	var b bytes.Buffer
	w := lz4.NewWriter(&b)
	w.Write(data)
	w.Close()
	return b.Bytes()
}

// LZ4 decompress
func decompressLZ4(data []byte) ([]byte, error) {
	r := lz4.NewReader(bytes.NewReader(data))
	var out bytes.Buffer
	_, err := out.ReadFrom(r)
	return out.Bytes(), err
}
