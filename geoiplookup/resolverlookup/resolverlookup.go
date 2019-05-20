// Package resolverlookup discovers the resolver's IP
package resolverlookup

import (
	"context"
	"net"
)

// Do performs the lookup of the resolver's IP.
func Do(ctx context.Context, resolver *net.Resolver) ([]string, error) {
	return resolver.LookupHost(ctx, "whoami.akamai.net")
}
