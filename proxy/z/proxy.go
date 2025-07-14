package z

type ZProxy struct {
	Http      bool
	Port      int
	Host      string
	Transport string
}

func (z *ZProxy) Inbound() error {
	// Implement the logic for handling inbound connections
	return nil
}

func (z *ZProxy) Outbound() error {
	// Implement the logic for handling outbound connections
	return nil
}
