package transport

import "net"

type ITransport interface {
	// Listen starts listening for incoming connections on the specified port.
	Listen(port int, handler func(con net.Conn))

	// Connect establishes a connection to the specified IP and port.
	Connect(ip string, port int) net.Conn

	// Send sends data over the established connection.
	Send(con net.Conn, server net.Conn, s func([]byte) []byte, r func([]byte) []byte)

	// IsAlive
	IsAlive(con net.Conn) bool
}
