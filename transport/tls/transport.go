package tls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type TLS struct {
}

func (r *TLS) Listen() *net.Listener {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			fmt.Println("📡 SNI Received:", chi.ServerName)
			return nil, nil // Use default
		},
	}
	ln, err := tls.Listen("tcp", ":4433", config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✅ TLS Server listening on :4433")
	return &ln
}
func (r *TLS) Handle(ln net.Listener, handle func(conn net.Conn)) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("❌ Accept failed:", err)
			continue
		}
		go handle(conn)
	}
}

func (r *TLS) Connect(sni string, address string) *tls.Conn {
	//address "localhost:4433"s
	conf := &tls.Config{
		ServerName:         sni, //"sni.fake.site",
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", address, conf)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	return conn
}
