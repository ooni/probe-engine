package netxlite

//
// Netx is a high-level structure that provides constructors for basic netxlite
// network operations using a custom model.UnderlyingNetwork.
//

import (
	"net"

	"github.com/ooni/probe-engine/pkg/model"
)

// Netx allows constructing netxlite data types using a specific [model.UnderlyingNetwork].
type Netx struct {
	// Underlying is the OPTIONAL [model.UnderlyingNetwork] to use. Leaving this field
	// nil makes this implementation functionally equivalent to netxlite top-level functions.
	Underlying model.UnderlyingNetwork
}

var _ model.MeasuringNetwork = &Netx{}

// MaybeCustomUnderlyingNetwork wraps the [model.UnderlyingNetwork] using a [*MaybeCustomUnderlyingNetwork].
func (netx *Netx) MaybeCustomUnderlyingNetwork() *MaybeCustomUnderlyingNetwork {
	return &MaybeCustomUnderlyingNetwork{netx.Underlying}
}

// ListenTCP creates a new listening TCP socket using the given address.
func (netx *Netx) ListenTCP(network string, addr *net.TCPAddr) (net.Listener, error) {
	return netx.MaybeCustomUnderlyingNetwork().Get().ListenTCP(network, addr)
}
