package wire

import "github.com/ooni/probe-engine/internal/oopsi/github.com/Psiphon-Labs/quic-go/internal/protocol"

// AckRange is an ACK range
type AckRange struct {
	Smallest protocol.PacketNumber
	Largest  protocol.PacketNumber
}

// Len returns the number of packets contained in this ACK range
func (r AckRange) Len() protocol.PacketNumber {
	return r.Largest - r.Smallest + 1
}
