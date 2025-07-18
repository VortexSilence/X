package z

import (
	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/transport"
	"net"
)

type Zed struct {
	Http      bool
	Port      int
	Host      string
	Transport string
}

func NewZed() *Zed {
	return &Zed{}
}

func (z *Zed) Inbound(c config.Inbound, o func() net.Conn) error {
	client := transport.New()
	client.Listen(c.Port, func(con net.Conn) {
		//handle Outbound
		client.Send(con, o(), z.Decode, func(b []byte) []byte {
			return b
		})
	})
	return nil
}

func (z *Zed) Outbound() net.Conn {
	client := transport.New()
	return client.Connect("127.0.0.1", 1080)
}

func (z *Zed) Encode(b []byte) []byte {
	return b
}

func (z *Zed) Decode(b []byte) []byte {
	return b
}
