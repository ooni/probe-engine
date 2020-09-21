package resolver

import (
	"context"

	"golang.org/x/net/idna"
)

// IDNAResolver is to support resolving Internationalized Domain Names.
// see RFC3492
type IDNAResolver struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost
func (r IDNAResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	// convert i18n names to punycode
	host, err := idna.ToASCII(hostname)
	if err != nil {
		return nil, err
	}
	return r.Resolver.LookupHost(ctx, host)
}

// Network returns the transport network (e.g., doh, dot)
func (r IDNAResolver) Network() string {
	return "idna"
}

// Address returns the upstream server address.
func (r IDNAResolver) Address() string {
	return ""
}

var _ Resolver = IDNAResolver{}
