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
			msg := e(buf[:n])
			if _, err := ser.Write(msg); err != nil {
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
			msg := d(buf[:n])
			if _, err := con.Write(msg); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
	}()
	wg.Wait()
}
