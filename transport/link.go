package transport

import (
	"core/features/inbound"
	"core/features/outbound"
	"core/transport/tcp"
)

func NewIn() inbound.IInbound {
	return &tcp.InTCP{}
}

func NewOu() outbound.IOutbound {
	return &tcp.OuTCP{}
}
