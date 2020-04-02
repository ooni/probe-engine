package resolver

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/modelx"
)

func testresolverquick(t *testing.T, resolver modelx.DNSResolver) {
	addrs, err := resolver.LookupHost(context.Background(), "dns.google.com")
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

func TestIntegrationDetectBogon(t *testing.T) {
	resolver := NewResolverSystem()
	ctx := modelx.WithMeasurementRoot(
		context.Background(), &modelx.MeasurementRoot{
			Beginning:   time.Now(),
			ErrDNSBogon: modelx.ErrDNSBogon,
			Handler:     handlers.NoHandler,
		})
	addrs, err := resolver.LookupHost(ctx, "localhost")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if err.Error() != modelx.FailureDNSBogonError {
		t.Fatal("not the error we expected to see")
	}
	if addrs != nil {
		t.Fatal("expected nil addrs here")
	}
}

func TestIntegrationNewResolverSystem(t *testing.T) {
	testresolverquick(t, NewResolverSystem())
}

func TestIntegrationNewResolverUDPAddress(t *testing.T) {
	testresolverquick(t, NewResolverUDP(
		new(net.Dialer), "8.8.8.8:53"))
}

func TestIntegrationNewResolverUDPDomain(t *testing.T) {
	testresolverquick(t, NewResolverUDP(
		new(net.Dialer), "dns.google.com:53"))
}

func TestIntegrationNewResolverTCPAddress(t *testing.T) {
	testresolverquick(t, NewResolverTCP(
		new(net.Dialer), "8.8.8.8:53"))
}

func TestIntegrationNewResolverTCPDomain(t *testing.T) {
	testresolverquick(t, NewResolverTCP(
		new(net.Dialer), "dns.google.com:53"))
}

func TestIntegrationNewResolverDoTAddress(t *testing.T) {
	testresolverquick(t, NewResolverTLS(
		&tlsdialer{}, "9.9.9.9:853"))
}

func TestIntegrationNewResolverDoTDomain(t *testing.T) {
	testresolverquick(t, NewResolverTLS(
		&tlsdialer{}, "dns.quad9.net:853"))
}

func TestIntegrationNewResolverDoH(t *testing.T) {
	testresolverquick(t, NewResolverHTTPS(
		http.DefaultClient, "https://cloudflare-dns.com/dns-query"))
}
