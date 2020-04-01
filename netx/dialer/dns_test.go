package dialer_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegrationDNSDialerDial(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	conn, err := dialer.DialContext(context.Background(), "tcp", "www.google.com:80")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDNSDialerDialAddress(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	conn, err := dialer.DialContext(context.Background(), "tcp", "8.8.8.8:853")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationDNSDialerNoPort(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	conn, err := dialer.DialContext(context.Background(), "tcp", "antani.ooni.io")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestIntegrationDNSDialerLookupFailure(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	conn, err := dialer.DialContext(context.Background(), "tcp", "antani.ooni.io:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestIntegrationDNSDialerDialTCPFailure(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	// The port is unreachable and filtered. The timeout is here
	// to make sure that we don't run for too much time.
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp", "ooni.io:12345")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}

func TestUnitReduceErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		result := dialer.ReduceErrors(nil)
		if result != nil {
			t.Fatal("wrong result")
		}
	})

	t.Run("single error", func(t *testing.T) {
		err := errors.New("mocked error")
		result := dialer.ReduceErrors([]error{err})
		if result != err {
			t.Fatal("wrong result")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		err1 := errors.New("mocked error #1")
		err2 := errors.New("mocked error #2")
		result := dialer.ReduceErrors([]error{err1, err2})
		if result.Error() != "mocked error #1" {
			t.Fatal("wrong result")
		}
	})

	t.Run("multiple errors with meaningful ones", func(t *testing.T) {
		err1 := errors.New("mocked error #1")
		err2 := &modelx.ErrWrapper{
			Failure: "unknown_error: antani",
		}
		err3 := &modelx.ErrWrapper{
			Failure: modelx.FailureConnectionRefused,
		}
		err4 := errors.New("mocked error #3")
		result := dialer.ReduceErrors([]error{err1, err2, err3, err4})
		if result.Error() != modelx.FailureConnectionRefused {
			t.Fatal("wrong result")
		}
	})
}

func TestIntegrationDNSDialerDivertLookupHost(t *testing.T) {
	dialer := dialer.DNSDialer{Dialer: new(net.Dialer), Resolver: new(net.Resolver)}
	failure := errors.New("mocked error")
	root := &modelx.MeasurementRoot{
		Beginning: time.Now(),
		Handler:   handlers.NoHandler,
		LookupHost: func(ctx context.Context, hostname string) ([]string, error) {
			return nil, failure
		},
	}
	ctx := modelx.WithMeasurementRoot(context.Background(), root)
	conn, err := dialer.DialContext(ctx, "tcp", "google.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !errors.Is(err, failure) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected a nil conn here")
	}
}
