package mmdblookup_test

import (
	"testing"

	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
)

const (
	asnDBPath     = "../../testdata/asn.mmdb"
	countryDBPath = "../../testdata/country.mmdb"
	ipAddr        = "35.204.49.125"
)

func TestLookupProbeASN(t *testing.T) {
	asn, org, err := mmdblookup.LookupASN(asnDBPath, ipAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(asn)
	t.Log(org)
}

func TestLookupProbeCC(t *testing.T) {
	cc, err := mmdblookup.LookupCC(countryDBPath, ipAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cc)
}
