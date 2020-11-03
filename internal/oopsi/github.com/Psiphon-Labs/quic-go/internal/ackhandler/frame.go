package ackhandler

import "github.com/ooni/probe-engine/internal/oopsi/github.com/Psiphon-Labs/quic-go/internal/wire"

type Frame struct {
	wire.Frame // nil if the frame has already been acknowledged in another packet
	OnLost     func(wire.Frame)
	OnAcked    func(wire.Frame)
}
