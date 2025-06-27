package tcp

import (
	"bufio"
	"core/transport/pipe"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
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
		t.sClient(con, proto, port)
	} else {
		t.sServer(con, proto, port)
	}
}
func (t *OuTCP) sClient(clientConn net.Conn, proto string, port int) {
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
			msg := append([]byte(proto+":"), buf[:n]...)
			// start encode

			// start decode
			wrapped := pipe.HandlePipe(msg)
			if _, err := serverConn.Write(wrapped); err != nil {
				log.Printf("Write to server error: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		reader := bufio.NewReader(serverConn)
		for {
			resp, err := http.ReadResponse(reader, nil)
			if err != nil {
				log.Printf("Read HTTP response error: %v", err)
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Read response body error: %v", err)
				return
			}
			if _, err := clientConn.Write(body); err != nil {
				log.Printf("Write to client error: %v", err)
				return
			}
		}
	}()

	wg.Wait()
}

func (t *OuTCP) sServer(tunnelConn net.Conn, proto string, port int) {
	defer tunnelConn.Close()

	reader := bufio.NewReader(tunnelConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("HTTP read error: %v", err)
		return
	}

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Read payload error: %v", err)
		return
	}

	var protocol string
	var realData []byte
	if parts := strings.SplitN(string(payload), ":", 2); len(parts) == 2 {
		protocol = parts[0]
		realData = []byte(parts[1])
	} else {
		log.Printf("Invalid protocol prefix in payload")
		return
	}

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
			response := fmt.Sprintf(
				"HTTP/1.1 200 OK\r\n"+
					"Host: %s\r\n"+
					"User-Agent: %s\r\n"+
					"Content-Type: application/octet-stream\r\n"+
					"Content-Length: %d\r\n\r\n%s",
				"pashmak.com", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", n+len(protocol)+1, protocol+":"+string(buf[:n]))

			if _, err := tunnelConn.Write([]byte(response)); err != nil {
				log.Printf("Write to tunnel error: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		reader := bufio.NewReader(tunnelConn)
		for {
			req, err := http.ReadRequest(reader)
			if err != nil {
				if err != io.EOF {
					log.Printf("HTTP read error: %v", err)
				}
				return
			}

			payload, err := io.ReadAll(req.Body)
			if err != nil {
				log.Printf("Read payload error: %v", err)
				return
			}

			if _, err := targetConn.Write(payload); err != nil {
				log.Printf("Write to target error: %v", err)
				return
			}
		}
	}()

	wg.Wait()
}
