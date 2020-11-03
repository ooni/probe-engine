package utils

import "github.com/ooni/probe-engine/internal/oopsi/github.com/Psiphon-Labs/quic-go/internal/protocol"

// ByteInterval is an interval from one ByteCount to the other
type ByteInterval struct {
	Start protocol.ByteCount
	End   protocol.ByteCount
}
