// Package dialer contains the dialer's API. The dialer defined
// in here implements basic DNS, but that is overridable.
package dialer

import (
	"crypto/tls"

	"github.com/ooni/probe-engine/netx/modelx"
)

// New creates a new modelx.Dialer
func New(resolver modelx.DNSResolver, dialer modelx.Dialer) *DNSDialer {
	return NewDNSDialer(resolver, dialer)
}

// NewTLS creates a new modelx.TLSDialer
func NewTLS(dialer modelx.Dialer, config *tls.Config) *TLSDialer {
	return NewTLSDialer(dialer, config)
}
