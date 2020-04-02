package brokenresolver

import (
	"context"
	"testing"
)

func TestLookupAddr(t *testing.T) {
	client := NewBrokenResolver()
	names, err := client.LookupAddr(context.Background(), "8.8.8.8")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if names != nil {
		t.Fatal("expected nil here")
	}
}

func TestLookupCNAME(t *testing.T) {
	client := NewBrokenResolver()
	cname, err := client.LookupCNAME(context.Background(), "www.ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if cname != "" {
		t.Fatal("expected empty string here")
	}
}

func TestLookupHost(t *testing.T) {
	client := NewBrokenResolver()
	addrs, err := client.LookupHost(context.Background(), "www.google.com")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if addrs != nil {
		t.Fatal("expected nil here")
	}
}

func TestLookupMX(t *testing.T) {
	client := NewBrokenResolver()
	records, err := client.LookupMX(context.Background(), "ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if records != nil {
		t.Fatal("expected nil here")
	}
}

func TestLookupNS(t *testing.T) {
	client := NewBrokenResolver()
	records, err := client.LookupNS(context.Background(), "ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if records != nil {
		t.Fatal("expected nil here")
	}
}
