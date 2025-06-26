package handler

import (
	"core/config"
	"core/transport"
	"net"
)

func Handle() {
	//TODO handler port later
	c := config.Get()
	out := transport.NewOu()
	transport.NewIn().Listen(c.Port, func(con net.Conn) {
		out.Send(con, "tcp", c.ToPort, c.Mode)
	})
}
