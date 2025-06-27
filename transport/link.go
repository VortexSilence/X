package transport

import (
	"github.com/VortexSilence/X/features/inbound"
	"github.com/VortexSilence/X/features/outbound"
	"github.com/VortexSilence/X/transport/tcp"
)

func NewIn() inbound.IInbound {
	return &tcp.InTCP{}
}

func NewOu() outbound.IOutbound {
	return &tcp.OuTCP{}
}
