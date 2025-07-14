package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/VortexSilence/X/transport/upgrade"
)

type TCP struct {
}

func (t *TCP) Listen(port int, h func(con net.Conn)) {
	tcpLn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("TCP listen error: %v", err)
	}
	fmt.Printf("TCP listening on port %d\n", port)
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

func (t *TCP) ListenHTTP(port int, h func(con net.Conn)) {
	upgrade.NewUpgrade().Handle(port, func(w net.Conn) {
		h(w)
	})
}

func (t *TCP) IsAlive(conn net.Conn) bool {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	testByte := []byte{0}
	_, err := conn.Read(testByte)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}
		return false
	}
	return true
}

func (r *TCP) Connect(ip string, port int) net.Conn {
	//addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port))
	//if err != nil {
	//	log.Printf("TCP resolve error: %v", err)
	//}
	//conn, err := net.DialTCP("tcp", nil, addr)
	//if err != nil {
	//	log.Printf("TCP dial error: %v", err)
	//}
	//err = conn.SetKeepAlive(true)
	//if err != nil {
	//	log.Printf("TCP set keep-alive error: %v", err)
	//}
	//log.Printf("TCP connected to %s:%d", ip, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Printf("TCP connect error: %v", err)
		return nil
	}

	return conn
}

func (t *TCP) Send(con net.Conn, ser net.Conn, e func(b []byte) []byte, d func(b []byte) []byte) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024*1024)
		for {
			n, err := con.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %v", err)
				}
				return
			}
			if n == 0 {
				continue
			}
			//var msg []byte
			//if encode {
			//	msg = pipe.HandlePipeEncoder(buf[:n])
			//} else {
			//	msg = pipe.HandlePipeDecoder(buf[:n])
			//}
			//msg := pipe.HandlePipeEncoder(buf[:n])
			if _, err := ser.Write(buf[:n]); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 1024*1024)
		for {
			n, err := ser.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %v", err)
				}
				return
			}
			if n == 0 {
				continue
			}
			//TODO: handle decode
			if _, err := con.Write(buf[:n]); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}()
	wg.Wait()
}

/*


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

*/
