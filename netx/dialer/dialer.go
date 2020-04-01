// Package dialer contains the dialer's API. The dialer defined
// in here implements basic DNS, but that is overridable.
package dialer

import (
	"crypto/tls"

	"github.com/ooni/probe-engine/netx/dialer/dnsdialer"
	"github.com/ooni/probe-engine/netx/dialer/tlsdialer"
	"github.com/ooni/probe-engine/netx/modelx"
)

// New creates a new modelx.Dialer
func New(resolver modelx.DNSResolver, dialer modelx.Dialer) *dnsdialer.Dialer {
	return dnsdialer.New(resolver, dialer)
}

// NewTLS creates a new modelx.TLSDialer
func NewTLS(dialer modelx.Dialer, config *tls.Config) *tlsdialer.TLSDialer {
	return tlsdialer.New(dialer, config)
}
