// Package mmdblookup performs probe ASN, CC, NetworkName lookups.
package mmdblookup

import (
	"net"

	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/oschwald/geoip2-golang"
)

// LookupASN maps the ip to the probe ASN and org using the
// MMDB database located at path, or returns an error. In case
// the IP is not valid, this function will fail with an error
// complaining that geoip2 was passed a nil IP.
func LookupASN(
	path, ip string, logger log.Logger,
) (asn uint, org string, err error) {
	asn, org = model.DefaultProbeASN, model.DefaultProbeNetworkName
	db, err := geoip2.Open(path)
	if err != nil {
		return
	}
	defer db.Close()
	record, err := db.ASN(net.ParseIP(ip))
	if err != nil {
		return
	}
	logger.Debugf("mmdblookup: ASN: %+v", record)
	asn = record.AutonomousSystemNumber
	if record.AutonomousSystemOrganization != "" {
		org = record.AutonomousSystemOrganization
	}
	return
}

// LookupCC is like LookupASN but for the country code.
func LookupCC(
	path, ip string, logger log.Logger,
) (cc string, err error) {
	cc = model.DefaultProbeCC
	db, err := geoip2.Open(path)
	if err != nil {
		return
	}
	defer db.Close()
	record, err := db.Country(net.ParseIP(ip))
	if err != nil {
		return
	}
	logger.Debugf("mmdblookup: Country: %+v", record)
	// With MaxMind DB we used record.RegisteredCountry.IsoCode but that does
	// not seem to work with the db-ip.com database. The record is empty, at
	// least for my own IP address in Italy. --Simone (2020-02-25)
	if record.Country.IsoCode != "" {
		cc = record.Country.IsoCode
	}
	return
}
