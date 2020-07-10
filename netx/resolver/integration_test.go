package resolver_test

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/resolver"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func testresolverquick(t *testing.T, reso resolver.Resolver) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	reso = resolver.LoggingResolver{Logger: log.Log, Resolver: reso}
	addrs, err := reso.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if addrs == nil {
		t.Fatal("expected non-nil addrs here")
	}
	var foundquad8 bool
	for _, addr := range addrs {
		if addr == "8.8.8.8" {
			foundquad8 = true
		}
	}
	if !foundquad8 {
		t.Fatal("did not find 8.8.8.8 in ouput")
	}
}

func TestIntegrationNewResolverSystem(t *testing.T) {
	testresolverquick(t, resolver.SystemResolver{})
}

func TestIntegrationNewResolverUDPAddress(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverUDP(new(net.Dialer), "8.8.8.8:53")))
}

func TestIntegrationNewResolverUDPDomain(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverUDP(new(net.Dialer), "dns.google.com:53")))
}

func TestIntegrationNewResolverTCPAddress(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverTCP(new(net.Dialer).DialContext, "8.8.8.8:53")))
}

func TestIntegrationNewResolverTCPDomain(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverTCP(new(net.Dialer).DialContext, "dns.google.com:53")))
}

func TestIntegrationNewResolverDoTAddress(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverTLS(resolver.DialTLSContext, "8.8.8.8:853")))
}

func TestIntegrationNewResolverDoTDomain(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverTLS(resolver.DialTLSContext, "dns.google.com:853")))
}

func TestIntegrationNewResolverDoH(t *testing.T) {
	testresolverquick(t, resolver.NewSerialResolver(
		resolver.NewDNSOverHTTPS(http.DefaultClient, "https://doh.powerdns.org/")))
}
