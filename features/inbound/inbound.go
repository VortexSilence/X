package inbound

import "net"

type IInbound interface {
	Listen(port int, h func(con net.Conn))
}

// func NewInbound() {
// 	//if client
// 	if config.Get().Mode == "client" {
// 		any.NewAny().Run(1000, func(c net.Listener) {
// 			//call to proxy
// 		})
// 	}
// }
