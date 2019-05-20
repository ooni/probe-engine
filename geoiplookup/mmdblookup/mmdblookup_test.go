package mmdblookup_test

import (
	"testing"

	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
)

func TestLookupProbeASN(t *testing.T) {
	asn, org, err := mmdblookup.LookupASN("../../testdata/asn.mmdb", "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(asn)
	t.Log(org)
}

func TestLookupProbeCC(t *testing.T) {
	cc, err := mmdblookup.LookupCC("../../testdata/country.mmdb", "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cc)
}
