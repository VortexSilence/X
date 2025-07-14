package transport

import (
	"github.com/VortexSilence/X/features/transport"
	"github.com/VortexSilence/X/transport/tcp"
)

func New() transport.ITransport {
	return &tcp.TCP{}
}
