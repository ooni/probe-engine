// Package model defines shared data structures and interfaces.
package model

import "fmt"

const (
	// DefaultProbeASN is the default probe ASN as number.
	DefaultProbeASN uint = 0

	// DefaultProbeCC is the default probe CC.
	DefaultProbeCC = "ZZ"

	// DefaultProbeIP is the default probe IP.
	DefaultProbeIP = "127.0.0.1"

	// DefaultProbeNetworkName is the default probe network name.
	DefaultProbeNetworkName = ""

	// DefaultResolverASN is the default resolver ASN.
	DefaultResolverASN uint = 0

	// DefaultResolverIP is the default resolver IP.
	DefaultResolverIP = "127.0.0.2"

	// DefaultResolverNetworkName is the default resolver network name.
	DefaultResolverNetworkName = ""
)

var (
	// DefaultProbeASNString is the default probe ASN as a string.
	DefaultProbeASNString = fmt.Sprintf("AS%d", DefaultProbeASN)

	// DefaultResolverASNString is the default resolver ASN as a string.
	DefaultResolverASNString = fmt.Sprintf("AS%d", DefaultResolverASN)
)
