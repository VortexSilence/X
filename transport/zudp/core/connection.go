// connection.go
package transport

import (
	"errors"
	"net"
	"time"

	congestion "github.com/VortexSilence/X/transport/zudp/pkg/cubic"
)

type Connection struct {
	conn       net.Conn
	congestion *congestion.Cubic
	// سایر فیلدها
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:       conn,
		congestion: congestion.NewCubic(),
	}
}

func (c *Connection) Write(data []byte) (int, error) {
	// بررسی پنجره ازدحام
	if c.congestion.AvailableWindow() < uint32(len(data)) {
		return 0, errors.New("")
	}

	n, err := c.conn.Write(data)
	if err == nil {
		c.congestion.OnPacketSent(uint32(n))
	}
	return n, err
}

func (c *Connection) handleAck(seqNum uint32, rtt time.Duration) {
	c.congestion.OnPacketAck(1460, rtt) // فرض می‌کنیم هر بسته 1460 بایت است
}

func (c *Connection) handleLoss(seqNum uint32) {
	c.congestion.OnPacketLost(1460)
}
