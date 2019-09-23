package mmdblookup_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/resources"
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

func TestLookupProbeASN(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := mmdblookup.LookupASN(asnDBPath, ipAddr, log.Log)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(asn)
	t.Log(org)
}

func TestLookupProbeASNInvalidFile(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := mmdblookup.LookupASN("/nonexistent", ipAddr, log.Log)
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

func TestLookupProbeASNInvalidIP(t *testing.T) {
	maybeFetchResources(t)
	asn, org, err := mmdblookup.LookupASN(asnDBPath, "xxx", log.Log)
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

func TestLookupProbeCC(t *testing.T) {
	maybeFetchResources(t)
	cc, err := mmdblookup.LookupCC(countryDBPath, ipAddr, log.Log)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cc)
}

func TestLookupProbeCCInvalidFile(t *testing.T) {
	maybeFetchResources(t)
	cc, err := mmdblookup.LookupCC("/nonexistent", ipAddr, log.Log)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cc != model.DefaultProbeCC {
		t.Fatal("expected an empty cc")
	}
}

func TestLookupProbeCCInvalidIP(t *testing.T) {
	maybeFetchResources(t)
	cc, err := mmdblookup.LookupCC(asnDBPath, "xxx", log.Log)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cc != model.DefaultProbeCC {
		t.Fatal("expected an empty cc")
	}
}
