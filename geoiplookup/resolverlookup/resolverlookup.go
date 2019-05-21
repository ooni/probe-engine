// Package resolverlookup discovers the resolver's IP
package resolverlookup

import (
	"context"
	"net"
)

// Do performs the lookup of the resolver's IP.
func Do(ctx context.Context, resolver *net.Resolver) (ips []string, err error) {
	for i := 0; i < 3; i++ {
		ips, err = resolver.LookupHost(ctx, "whoami.akamai.net")
	}
	return
}
