package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/VortexSilence/X/transport/pipe"
)

type InTCP struct {
}

type OuTCP struct {
}

func (t *InTCP) Listen(port int, h func(con net.Conn)) {
	tcpLn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("TCP listen error: %v", err)
	}
	func() {
		for {
			conn, err := tcpLn.Accept()
			if err != nil {
				log.Printf("TCP accept error: %v", err)
				continue
			}
			go func() {
				h(conn)
			}()
		}
	}()
	//TODO: later
	// t.handleUDPConnections(udpLn)
}

func (t *OuTCP) Send(con net.Conn, proto string, port int, mode string) {
	if mode == "client" {
		t.sClient(con, port)
	} else {
		t.sServer(con, port)
	}
}
func (t *OuTCP) sClient(clientConn net.Conn, port int) {
	p := pipe.HandlePipe()
	serverConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Printf("Server connection error: %v", err)
		return
	}
	defer serverConn.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read from client error: %v", err)
				}
				return
			}
			// msg := append([]byte(proto+":"), buf[:n]...)
			s := pipe.HandlePipeEncoder(buf[:n])
			msg := p.Wrap(s, "proto")
			if _, err := serverConn.Write(msg); err != nil {
				log.Printf("Write to server error: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := serverConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read from server error: %v", err)
				}
				return
			}
			_, decoded, err := p.UnwrapResponse(buf[:n])
			if err != nil {
				log.Printf("Decode error: %v", err)
				return
			}
			if _, err := clientConn.Write(decoded); err != nil {
				log.Printf("Write to client error: %v", err)
				return
			}

		}
	}()

	wg.Wait()
}

func (t *OuTCP) sServer(tunnelConn net.Conn, port int) {
	p := pipe.HandlePipe()
	defer tunnelConn.Close()
	buf := make([]byte, 32*1024)
	n, err := tunnelConn.Read(buf)
	if err != nil {
		if err != io.EOF {
			log.Printf("Read from server error: %v", err)
		}
		return
	}
	protocol, realData, err := p.Unwrap(buf[:n])
	if err != nil {
		log.Printf("Unwrap error: %v", err)
		return
	}
	realData = pipe.HandlePipeEncoder(realData)
	fmt.Println(protocol)
	targetConn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("Target connection error: %v", err)
		return
	}
	defer targetConn.Close()

	if _, err := targetConn.Write(realData); err != nil {
		log.Printf("Write to target error: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := targetConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read from target error: %v", err)
				}
				return
			}
			wrappedResp := p.WrapResponse(buf[:n], "proto", 200)
			if _, err := tunnelConn.Write([]byte(wrappedResp)); err != nil {
				log.Printf("Write to tunnel error: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 32*1024)
		for {
			n, err := tunnelConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read from tunnel error: %v", err)
				}
				return
			}

			if _, err := targetConn.Write(buf[:n]); err != nil {
				log.Printf("Write to target error: %v", err)
				return
			}
		}
	}()

	wg.Wait()
}
