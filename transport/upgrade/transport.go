package upgrade

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
)

type Upgrade struct {
}

// p - protocol name tcp udp ws quic grpc
func (u *Upgrade) Upgrade(conn net.Conn, p string) bool {
	fmt.Fprintf(conn,
		"GET /tcp HTTP/1.1\r\n"+
			"Host: localhost\r\n"+
			"Upgrade: %s\r\n"+
			"Connection: Upgrade\r\n"+
			"\r\n", p)
	_, err := bufio.NewReader(conn).ReadString('\n')
	return err != nil
}

// handle and hijack http request to raw tcp connection
// TODO add ws and tcp and other protocols
// TODO add quic support for udp HTTP/3
func (u *Upgrade) Handle(handler func(conn net.Conn, rw *bufio.ReadWriter)) {
	http.HandleFunc("/tcp", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Incoming HTTP request...")
		if r.Header.Get("Upgrade") != "tcp" {
			http.Error(w, "Upgrade required", http.StatusUpgradeRequired)
			return
		}
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, "Hijack failed", http.StatusInternalServerError)
			return
		}
		log.Println("Connection upgraded, switching to raw QUIC mode...")
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\n")
		fmt.Fprintf(conn, "Upgrade: tcp\r\n")
		fmt.Fprintf(conn, "Connection: Upgrade\r\n\r\n")
		//now move to tcp with hijack http :D
		handler(conn, rw)
	})
	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
