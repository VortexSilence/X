package outbound

import "net"

type IOutbound interface {
	Send(clientConn net.Conn, proto string, port int, mode string)
}
