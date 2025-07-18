package freedom

import (
	"log"
	"net"
	"os"

	"github.com/VortexSilence/X/transport"
	"github.com/things-go/go-socks5"
)

type Freedom struct {
}

func NewFreedom() *Freedom {
	return &Freedom{}
}
func (f *Freedom) Listen() error {
	server := socks5.NewServer(
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
	)

	if err := server.ListenAndServe("tcp", ":8099"); err != nil {
		return err
	}
	return nil
}

func (f *Freedom) GetConnection() net.Conn {
	return transport.New().Connect("127.0.0.1", 8099)
}
