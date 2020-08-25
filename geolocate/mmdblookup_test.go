package geolocate_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geolocate"
	"github.com/ooni/probe-engine/resources"
	"github.com/ooni/probe-engine/model"
)

const (
	asnDBPath     = "../../testdata/asn.mmdb"
	countryDBPath = "../../testdata/country.mmdb"
	ipAddr        = "35.204.49.125"
)

func maybeFetchResources(t *testing.T) {
	c := &resources.Client{
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
		WorkDir:    "../../testdata/",
	}
	if err := c.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestLookupASN(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := geolocate.LookupASN(asnDBPath, ipAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(asn)
	t.Log(org)
}

func TestLookupASNInvalidFile(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := geolocate.LookupASN("/nonexistent", ipAddr)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if asn != model.DefaultProbeASN {
		t.Fatal("expected a zero ASN")
	}
	if org != model.DefaultProbeNetworkName {
		t.Fatal("expected an empty org")
	}
}

func TestLookupASNInvalidIP(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := geolocate.LookupASN(asnDBPath, "xxx")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if asn != model.DefaultProbeASN {
		t.Fatal("expected a zero ASN")
	}
	if org != model.DefaultProbeNetworkName {
		t.Fatal("expected an empty org")
	}
}

func TestLookupCC(t *testing.T) {
	maybeFetchResources(t)
	cc, err := geolocate.LookupCC(countryDBPath, ipAddr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cc)
}

func TestLookupCCInvalidFile(t *testing.T) {
	maybeFetchResources(t)
	cc, err := geolocate.LookupCC("/nonexistent", ipAddr)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cc != model.DefaultProbeCC {
		t.Fatal("expected an empty cc")
	}
}

func TestLookupCCInvalidIP(t *testing.T) {
	maybeFetchResources(t)
	cc, err := geolocate.LookupCC(asnDBPath, "xxx")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cc != model.DefaultProbeCC {
		t.Fatal("expected an empty cc")
	}
}
