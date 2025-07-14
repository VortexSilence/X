package rpc

import (
	"fmt"
	"net"
	"net/rpc"
)

type InRPC struct {
}

type OuRPC struct {
}

type Packet struct{}

type Args struct {
	A, B int
}

func (a *Packet) Multiply(args Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *InRPC) Listen() {
	arith := new(Packet)
	rpc.Register(arith)

	l, _ := net.Listen("tcp", ":1234")
	defer l.Close()
	fmt.Println("RPC server listening on port 1234")
	rpc.Accept(l)
}

func (t *OuRPC) Send() {

}
