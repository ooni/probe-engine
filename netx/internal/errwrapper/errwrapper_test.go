package errwrapper

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"syscall"
	"testing"

	"github.com/ooni/probe-engine/netx/modelx"
)

func TestMaybeBuildFactory(t *testing.T) {
	err := SafeErrWrapperBuilder{
		ConnID:        1,
		DialID:        10,
		Error:         errors.New("mocked error"),
		TransactionID: 100,
	}.MaybeBuild()
	var target *modelx.ErrWrapper
	if errors.As(err, &target) == false {
		t.Fatal("not the expected error type")
	}
	if target.ConnID != 1 {
		t.Fatal("wrong ConnID")
	}
	if target.DialID != 10 {
		t.Fatal("wrong DialID")
	}
	if target.Failure != "unknown_failure: mocked error" {
		t.Fatal("the failure string is wrong")
	}
	if target.TransactionID != 100 {
		t.Fatal("the transactionID is wrong")
	}
	if target.WrappedErr.Error() != "mocked error" {
		t.Fatal("the wrapped error is wrong")
	}
}

func TestToFailureString(t *testing.T) {
	t.Run("for already wrapped error", func(t *testing.T) {
		err := SafeErrWrapperBuilder{Error: io.EOF}.MaybeBuild()
		if toFailureString(err) != "eof_error" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for modelx.ErrDNSBogon", func(t *testing.T) {
		if toFailureString(modelx.ErrDNSBogon) != "dns_bogon_error" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for x509.HostnameError", func(t *testing.T) {
		var err x509.HostnameError
		if toFailureString(err) != "ssl_invalid_hostname" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for x509.UnknownAuthorityError", func(t *testing.T) {
		var err x509.UnknownAuthorityError
		if toFailureString(err) != "ssl_unknown_authority" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for x509.CertificateInvalidError", func(t *testing.T) {
		var err x509.CertificateInvalidError
		if toFailureString(err) != "ssl_invalid_certificate" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for EOF", func(t *testing.T) {
		if toFailureString(io.EOF) != "eof_error" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for connection_refused", func(t *testing.T) {
		if toFailureString(syscall.ECONNREFUSED) != "connection_refused" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for connection_reset", func(t *testing.T) {
		if toFailureString(syscall.ECONNRESET) != "connection_reset" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for context deadline expired", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()
		<-ctx.Done()
		if toFailureString(ctx.Err()) != "generic_timeout_error" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for i/o error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", "www.google.com:80")
		if err == nil {
			t.Fatal("expected an error here")
		}
		if conn != nil {
			t.Fatal("expected nil connection here")
		}
		if toFailureString(err) != "generic_timeout_error" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for TLS handshake timeout error", func(t *testing.T) {
		err := errors.New("net/http: TLS handshake timeout")
		if toFailureString(err) != "generic_timeout_error" {
			t.Fatal("unexpected results")
		}
	})
	t.Run("for no such host", func(t *testing.T) {
		if toFailureString(&net.DNSError{
			Err: "no such host",
		}) != "dns_nxdomain_error" {
			t.Fatal("unexpected results")
		}
	})
}

func TestUnitToOperationString(t *testing.T) {
	t.Run("for connect", func(t *testing.T) {
		// You're doing HTTP and connect fails. You want to know
		// that connect failed not that HTTP failed.
		err := &modelx.ErrWrapper{Operation: "connect"}
		if toOperationString(err, "http_round_trip") != "connect" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for http_round_trip", func(t *testing.T) {
		// You're doing DoH and something fails inside HTTP. You want
		// to know about the internal HTTP error, not resolve.
		err := &modelx.ErrWrapper{Operation: "http_round_trip"}
		if toOperationString(err, "resolve") != "http_round_trip" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for resolve", func(t *testing.T) {
		// You're doing HTTP and the DNS fails. You want to
		// know that resolve failed.
		err := &modelx.ErrWrapper{Operation: "resolve"}
		if toOperationString(err, "http_round_trip") != "resolve" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for tls_handshake", func(t *testing.T) {
		// You're doing HTTP and the TLS handshake fails. You want
		// to know about a TLS handshake error.
		err := &modelx.ErrWrapper{Operation: "tls_handshake"}
		if toOperationString(err, "http_round_trip") != "tls_handshake" {
			t.Fatal("unexpected result")
		}
	})
	t.Run("for minor operation", func(t *testing.T) {
		// You just noticed that TLS handshake failed and you
		// have a child error telling you that read failed. Here
		// you want to know about a TLS handshake error.
		err := &modelx.ErrWrapper{Operation: "read"}
		if toOperationString(err, "tls_handshake") != "tls_handshake" {
			t.Fatal("unexpected result")
		}
	})
}
