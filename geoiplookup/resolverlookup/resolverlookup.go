// Package resolverlookup discovers the resolver's IP
package resolverlookup

import (
	"context"
	"net"

	"github.com/ooni/probe-engine/httpx/retryx"
)

// Do performs the lookup of the resolver's IP.
func Do(ctx context.Context, resolver *net.Resolver) (ips []string, err error) {
	err = retryx.Do(ctx, func() error {
		ips, err = resolver.LookupHost(ctx, "whoami.akamai.net")
		return err
	})
	return
}
