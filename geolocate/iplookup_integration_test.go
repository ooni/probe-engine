package geolocate_test

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/geolocate"
	"github.com/ooni/probe-engine/internal/httpheader"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestIPLookupWorksUsingAvast(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.AvastIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingEkiga(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.STUNEkigaIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingGoogle(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.STUNGoogleIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingIPConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.IPConfigIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingIPInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.IPInfoIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingUbuntu(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	ip, err := geolocate.UbuntuIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		httpheader.UserAgent(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}
