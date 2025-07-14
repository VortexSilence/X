package freedom

import (
	"github.com/things-go/go-socks5"
	"log"
	"os"
)

type Freedom struct {
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

func (f *Freedom) Send() {
}
