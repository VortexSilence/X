package tcp

import (
	"fmt"
	"net"
	"time"
)

type UDP struct {
}

func (t *UDP) Listen() *net.UDPConn {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		panic(err)
	}
	defer con.Close()
	localAddr := con.LocalAddr().(*net.UDPAddr)
	fmt.Println("Client UDP listening on", localAddr.String())
	return con
}

func (t *UDP) Read(con net.UDPConn) []byte {
	buf := make([]byte, 1024)
	con.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, addr, err := con.ReadFromUDP(buf)
	if err != nil {
		return nil
	}
	fmt.Printf("Got UDP message from %s: %s\n", addr.String(), string(buf[:n]))
	return buf[:n]
}

func (t *UDP) Send(conn net.Conn, b []byte) {
	conn.Write(b)

}

func (t *UDP) Connect(addp string) *net.Conn {
	con, err := net.Dial("udp", addp)
	if err != nil {
		return nil
	}
	defer con.Close()
	return &con
}
