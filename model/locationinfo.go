package model

// LocationInfo contains location information
type LocationInfo struct {
	// ASN is the autonomous system number
	ASN uint

	// CountryCode is the country code
	CountryCode string

	// DidResolverLookup indicates whether we did a resolver lookup.
	DidResolverLookup bool

	// NetworkName is the network name
	NetworkName string

	// IP is the probe IP
	ProbeIP string

	// ResolverASN is the resolver ASN
	ResolverASN uint

	// ResolverIP is the resolver IP
	ResolverIP string

	// ResolverNetworkName is the resolver network name
	ResolverNetworkName string
}
