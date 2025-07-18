package handler

import (
	"net"

	"time"

	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/proxy/freedom"
	"github.com/VortexSilence/X/proxy/z"
	"github.com/VortexSilence/X/transport"
)

func Handle() {
	config := config.Get()
	outbound := ""
	var out net.Conn
	for _, e := range config.Outbounds {
		if e.Protocol == "freedom" {
			outbound = "freedom"
			f := freedom.NewFreedom()
			go f.Listen()
			time.Sleep(5 * time.Second)
			out = f.GetConnection()
		}
		if e.Protocol == "zed" {
			outbound = "zed"

		}
	}
	for _, e := range config.Inbounds {
		client := transport.New()
		if e.Protocol == "any" {
			client.Listen(e.Port, func(con net.Conn) {
				if outbound == "freedom" {
					out = transport.New().Connect("127.0.0.1", 8099)
				}
				if outbound == "zed" {
					out = z.NewZed().Outbound()
				}
				go client.Send(con, out, z.NewZed().Encode, func(b []byte) []byte {
					return b
				})
			})
		}
		if e.Protocol == "zed" {
			z.NewZed().Inbound(e, func() net.Conn {
				if outbound == "freedom" {
					return freedom.NewFreedom().GetConnection()
				}
				return nil
			})
		}

	}
}
