package socks

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	socksVersion       = 0x05
	authNoAccept       = 0xff
	authUsernamePasswd = 0x02
	cmdConnect         = 0x01
	addrTypeIPv4       = 0x01
	addrTypeDomain     = 0x03
	addrTypeIPv6       = 0x04
)

var Users = map[string]string{
	"user": "pass",
}

type Socks struct {
}

func (s *Socks) Handle(conn net.Conn) []byte {
	defer conn.Close()
	buf := make([]byte, 1024*1024)

	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		log.Println("Handshake error:", err)
		return nil
	}
	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		log.Println("Read methods error:", err)
		return nil
	}

	supported := false
	for _, m := range buf[:nMethods] {
		if m == authUsernamePasswd {
			supported = true
			break
		}
	}
	if !supported {
		conn.Write([]byte{socksVersion, authNoAccept})
		return nil
	}
	conn.Write([]byte{socksVersion, authUsernamePasswd})

	// auth
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		log.Println("Auth header read error:", err)
		return nil
	}
	ulen := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:ulen]); err != nil {
		log.Println("Username read error:", err)
		return nil
	}
	username := string(buf[:ulen])

	if _, err := io.ReadFull(conn, buf[:1]); err != nil {
		log.Println("Password length read error:", err)
		return nil
	}
	plen := int(buf[0])
	if _, err := io.ReadFull(conn, buf[:plen]); err != nil {
		log.Println("Password read error:", err)
		return nil
	}
	password := string(buf[:plen])

	if pw, ok := Users[username]; !ok || pw != password {
		conn.Write([]byte{0x01, 0x01})
		return nil
	}
	conn.Write([]byte{0x01, 0x00})

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		log.Println("Request header error:", err)
		return nil
	}
	if buf[1] != cmdConnect {
		log.Println("Unsupported command:", buf[1])
		conn.Write([]byte{socksVersion, 0x07})
		return nil
	}

	var destAddr string
	switch buf[3] {
	case addrTypeIPv4:
		ip := make([]byte, 4)
		port := make([]byte, 2)
		io.ReadFull(conn, ip)
		io.ReadFull(conn, port)
		destAddr = fmt.Sprintf("%s:%d", net.IP(ip), binary.BigEndian.Uint16(port))
	case addrTypeDomain:
		io.ReadFull(conn, buf[:1])
		dlen := int(buf[0])
		io.ReadFull(conn, buf[:dlen])
		domain := string(buf[:dlen])
		io.ReadFull(conn, buf[:2])
		port := binary.BigEndian.Uint16(buf[:2])
		destAddr = fmt.Sprintf("%s:%d", domain, port)
	case addrTypeIPv6:
		ip := make([]byte, 16)
		port := make([]byte, 2)
		io.ReadFull(conn, ip)
		io.ReadFull(conn, port)
		destAddr = fmt.Sprintf("[%s]:%d", net.IP(ip), binary.BigEndian.Uint16(port))
	default:
		log.Println("Unknown address type:", buf[3])
		return nil
	}

	log.Printf("User [%s] requested %s", username, destAddr)

	//conn.Write([]byte{
	//	0x05, 0x00, 0x00, 0x01,
	//	0, 0, 0, 0,
	//		0, 0,
	//	})

	// backend, err := net.Dial("tcp", "127.0.0.1:8099")
	//
	//	if err != nil {
	//		log.Println("Failed to connect to backend:", err)
	//		return
	//	}
	//
	// defer backend.Close()
	//
	// backend.Write([]byte(destAddr + "\n"))
	//
	// go io.Copy(backend, conn)
	// io.Copy(conn, backend)
	return buf
}
