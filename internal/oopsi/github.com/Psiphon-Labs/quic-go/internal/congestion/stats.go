package congestion

import "github.com/ooni/probe-engine/internal/oopsi/github.com/Psiphon-Labs/quic-go/internal/protocol"

type connectionStats struct {
	slowstartPacketsLost protocol.PacketNumber
	slowstartBytesLost   protocol.ByteCount
}
