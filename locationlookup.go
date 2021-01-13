package engine

import (
	"context"
	"fmt"

	"github.com/ooni/probe-engine/model"
)

// LocationLookupResourceUpdater updates resources before
// performing the real location lookup.
type LocationLookupResourceUpdater interface {
	MaybeUpdateResources(ctx context.Context) error
}

// LocationLookupProbeIPLookupper lookups the Probe IP.
type LocationLookupProbeIPLookupper interface {
	LookupProbeIP(ctx context.Context) (addr string, err error)
}

// LocationLookupASNLookupper lookups the Probe ASN.
type LocationLookupASNLookupper interface {
	LookupASN(path string, ip string) (asn uint, network string, err error)
}

// LocationLookupCountryLookupper lookups the Probe CC.
type LocationLookupCountryLookupper interface {
	LookupCC(path string, ip string) (cc string, err error)
}

// LocationLookupResolverIPLookupper lookups the resolver IP.
type LocationLookupResolverIPLookupper interface {
	LookupResolverIP(ctx context.Context) (addr string, err error)
}

// LocationLookupsPathsProvider provides paths.
type LocationLookupsPathsProvider interface {
	ASNDatabasePath() string
	CountryDatabasePath() string
}

// LocationLookup performs a location lookup. Usually every field is
// initialised to a Session. You may wanna do otherwise in tests.
type LocationLookup struct {
	CountryLookupper     LocationLookupCountryLookupper
	EnableResolverLookup bool
	PathsProvider        LocationLookupsPathsProvider
	ProbeIPLookupper     LocationLookupProbeIPLookupper
	ProbeASNLookupper    LocationLookupASNLookupper
	ResolverASNLookupper LocationLookupASNLookupper
	ResolverIPLookupper  LocationLookupResolverIPLookupper
	ResourceUpdater      LocationLookupResourceUpdater
}

// Do performs a location lookup.
func (op LocationLookup) Do(ctx context.Context) (*model.LocationInfo, error) {
	var err error
	out := &model.LocationInfo{
		ASN:                 model.DefaultProbeASN,
		CountryCode:         model.DefaultProbeCC,
		NetworkName:         model.DefaultProbeNetworkName,
		ProbeIP:             model.DefaultProbeIP,
		ResolverASN:         model.DefaultResolverASN,
		ResolverIP:          model.DefaultResolverIP,
		ResolverNetworkName: model.DefaultResolverNetworkName,
	}
	if err := op.ResourceUpdater.MaybeUpdateResources(ctx); err != nil {
		return out, fmt.Errorf("MaybeUpdateResource failed: %w", err)
	}
	ip, err := op.ProbeIPLookupper.LookupProbeIP(ctx)
	if err != nil {
		return out, fmt.Errorf("lookupProbeIP failed: %w", err)
	}
	out.ProbeIP = ip
	asn, networkName, err := op.ProbeASNLookupper.LookupASN(
		op.PathsProvider.ASNDatabasePath(), out.ProbeIP)
	if err != nil {
		return out, fmt.Errorf("lookupASN failed: %w", err)
	}
	out.ASN = asn
	out.NetworkName = networkName
	cc, err := op.CountryLookupper.LookupCC(
		op.PathsProvider.CountryDatabasePath(), out.ProbeIP)
	if err != nil {
		return out, fmt.Errorf("lookupProbeCC failed: %w", err)
	}
	out.CountryCode = cc
	if op.EnableResolverLookup {
		out.DidResolverLookup = true
		// Note: ignoring the result of lookupResolverIP and lookupASN
		// here is intentional. We don't want this (~minor) failure
		// to influence the result of the overall lookup. Another design
		// here could be that of retrying the operation N times?
		resolverIP, err := op.ResolverIPLookupper.LookupResolverIP(ctx)
		if err != nil {
			return out, nil
		}
		out.ResolverIP = resolverIP
		resolverASN, resolverNetworkName, err := op.ResolverASNLookupper.LookupASN(
			op.PathsProvider.ASNDatabasePath(), out.ResolverIP,
		)
		if err != nil {
			return out, nil
		}
		out.ResolverASN = resolverASN
		out.ResolverNetworkName = resolverNetworkName
	}
	return out, nil
}
