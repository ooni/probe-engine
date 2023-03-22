package netx

//
// TLSDialer from Config.
//

import (
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

// NewTLSDialer creates a new TLSDialer from the specified config.
func NewTLSDialer(config Config) model.TLSDialer {
	if config.Dialer == nil {
		config.Dialer = NewDialer(config)
	}
	logger := model.ValidLoggerOrDefault(config.Logger)
	thx := netxlite.NewTLSHandshakerStdlib(logger)
	thx = config.Saver.WrapTLSHandshaker(thx) // WAI even when config.Saver is nil
	tlsConfig := netxlite.ClonedTLSConfigOrNewEmptyConfig(config.TLSConfig)
	return netxlite.NewTLSDialerWithConfig(config.Dialer, thx, tlsConfig)
}
