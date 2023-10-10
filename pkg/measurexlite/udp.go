package measurexlite

import "github.com/ooni/probe-engine/pkg/model"

func (tx *Trace) NewUDPListener() model.UDPListener {
	return tx.Netx.NewUDPListener()
}
