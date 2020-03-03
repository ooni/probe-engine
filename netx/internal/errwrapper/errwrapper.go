// Package errwrapper contains our error wrapper
package errwrapper

import (
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"github.com/ooni/probe-engine/netx/modelx"
)

// SafeErrWrapperBuilder contains a builder for modelx.ErrWrapper that
// is safe, i.e., behaves correctly when the error is nil.
type SafeErrWrapperBuilder struct {
	// ConnID is the connection ID, if any
	ConnID int64

	// DialID is the dial ID, if any
	DialID int64

	// Error is the error, if any
	Error error

	// Operation is the operation that failed
	Operation string

	// TransactionID is the transaction ID, if any
	TransactionID int64
}

// MaybeBuild builds a new modelx.ErrWrapper, if b.Error is not nil, and returns
// a nil error value, instead, if b.Error is nil.
func (b SafeErrWrapperBuilder) MaybeBuild() (err error) {
	if b.Error != nil {
		err = &modelx.ErrWrapper{
			ConnID:        b.ConnID,
			DialID:        b.DialID,
			Failure:       toFailureString(b.Error),
			Operation:     toOperationString(b.Error, b.Operation),
			TransactionID: b.TransactionID,
			WrappedErr:    b.Error,
		}
	}
	return
}

func toFailureString(err error) string {
	// The list returned here matches the values used by MK unless
	// explicitly noted otherwise with a comment.

	var errwrapper *modelx.ErrWrapper
	if errors.As(err, &errwrapper) {
		return errwrapper.Error() // we've already wrapped it
	}

	if errors.Is(err, modelx.ErrDNSBogon) {
		return "dns_bogon_error" // not in MK
	}

	var x509HostnameError x509.HostnameError
	if errors.As(err, &x509HostnameError) {
		// Test case: https://wrong.host.badssl.com/
		return "ssl_invalid_hostname"
	}
	var x509UnknownAuthorityError x509.UnknownAuthorityError
	if errors.As(err, &x509UnknownAuthorityError) {
		// Test case: https://self-signed.badssl.com/. This error has
		// never been among the ones returned by MK.
		return "ssl_unknown_authority"
	}
	var x509CertificateInvalidError x509.CertificateInvalidError
	if errors.As(err, &x509CertificateInvalidError) {
		// Test case: https://expired.badssl.com/
		return "ssl_invalid_certificate"
	}

	s := err.Error()
	if strings.HasSuffix(s, "EOF") {
		return "eof_error"
	}
	if strings.HasSuffix(s, "connection refused") {
		return "connection_refused"
	}
	if strings.HasSuffix(s, "connection reset by peer") {
		return "connection_reset"
	}
	if strings.HasSuffix(s, "context deadline exceeded") {
		return "generic_timeout_error"
	}
	if strings.HasSuffix(s, "i/o timeout") {
		return "generic_timeout_error"
	}
	if strings.HasSuffix(s, "TLS handshake timeout") {
		return "generic_timeout_error"
	}
	if strings.HasSuffix(s, "no such host") {
		// This is dns_lookup_error in MK but such error is used as a
		// generic "hey, the lookup failed" error. Instead, this error
		// that we return here is significantly more specific.
		return "dns_nxdomain_error"
	}

	return fmt.Sprintf("unknown_failure: %s", s)
}

func toOperationString(err error, operation string) string {
	var errwrapper *modelx.ErrWrapper
	if errors.As(err, &errwrapper) {
		// Basically, as explained in modelx.ErrWrapper docs, let's
		// keep the child major operation, if any.
		if errwrapper.Operation == "connect" {
			return errwrapper.Operation
		}
		if errwrapper.Operation == "http_round_trip" {
			return errwrapper.Operation
		}
		if errwrapper.Operation == "resolve" {
			return errwrapper.Operation
		}
		if errwrapper.Operation == "tls_handshake" {
			return errwrapper.Operation
		}
		// FALLTHROUGH
	}
	return operation
}
