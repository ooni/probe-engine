package measurexlite

import "github.com/ooni/probe-engine/pkg/model"

// NewUDPListener implements model.Measuring Network.
func (tx *Trace) NewUDPListener() model.UDPListener {
	return tx.Netx.NewUDPListener()
}
