package http

import (
	"fmt"
	"net"
	"net/http"
)

type Http struct {
}

func (t *Http) Listen(port int, handler func(con net.Conn)) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//hijack and get con
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			return
		}
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\n")
		fmt.Fprintf(conn, "Upgrade: quicstep\r\n")
		fmt.Fprintf(conn, "Connection: Upgrade\r\n\r\n")
		handler(conn)
	})
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func (t *Http) Send() {
}
