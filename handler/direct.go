package handler

import (
	"fmt"
	"net"

	"time"

	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/proxy/freedom"
	"github.com/VortexSilence/X/transport"
)

func Handle() {
	config := config.Get()
	outbound := ""
	var out net.Conn
	for _, e := range config.Outbounds {
		if e.Protocol == "freedom" {
			outbound = "freedom"
			f := freedom.Freedom{}
			go f.Listen()
			time.Sleep(5 * time.Second)
			out = transport.New().Connect("127.0.0.1", 8099)
		}
	}
	for _, e := range config.Inbounds {
		client := transport.New()
		if e.Protocol == "any" {
			client.Listen(e.Port, func(con net.Conn) {
				if outbound == "freedom" {
					out = transport.New().Connect("127.0.1", 8099)
				}
				go client.Send(con, out, func(b []byte) []byte {
					return b
				}, func(b []byte) []byte {
					return b
				})
			})
		}
		if e.Protocol == "z" {

		}
		fmt.Println(e.Protocol)

	}
}
func Test() {
	fmt.Println("test")
}
