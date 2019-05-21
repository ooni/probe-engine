// Package mmdblookup performs probe ASN, CC, NetworkName lookups.
package mmdblookup

import (
	"net"

	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/oschwald/geoip2-golang"
)

// LookupASN maps the ip to the probe ASN and org using the
// MMDB database located at path, or returns an error. In case
// the IP is not valid, this function will fail with an error
// complaining that geoip2 was passed a nil IP.
func LookupASN(path, ip string) (asn uint, org string, err error) {
	asn, org = constants.DefaultProbeASNNumber, constants.DefaultProbeNetworkName
	db, err := geoip2.Open(path)
	if err != nil {
		return
	}
	defer db.Close()
	record, err := db.ASN(net.ParseIP(ip))
	if err != nil {
		return
	}
	asn = record.AutonomousSystemNumber
	if record.AutonomousSystemOrganization != "" {
		org = record.AutonomousSystemOrganization
	}
	return
}

// LookupCC is like LookupASN but for the country code.
func LookupCC(path, ip string) (cc string, err error) {
	cc = constants.DefaultProbeCC
	db, err := geoip2.Open(path)
	if err != nil {
		return
	}
	record, err := db.Country(net.ParseIP(ip))
	if err != nil {
		return
	}
	if record.RegisteredCountry.IsoCode != "" {
		cc = record.RegisteredCountry.IsoCode
	}
	return
}
