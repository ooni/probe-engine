// Package errorx contains error extensions
package errorx

import (
	"context"
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
		return modelx.FailureDNSBogonError // not in MK
	}
	if errors.Is(err, context.Canceled) {
		return modelx.FailureInterrupted
	}

	var x509HostnameError x509.HostnameError
	if errors.As(err, &x509HostnameError) {
		// Test case: https://wrong.host.badssl.com/
		return modelx.FailureSSLInvalidHostname
	}
	var x509UnknownAuthorityError x509.UnknownAuthorityError
	if errors.As(err, &x509UnknownAuthorityError) {
		// Test case: https://self-signed.badssl.com/. This error has
		// never been among the ones returned by MK.
		return modelx.FailureSSLUnknownAuthority
	}
	var x509CertificateInvalidError x509.CertificateInvalidError
	if errors.As(err, &x509CertificateInvalidError) {
		// Test case: https://expired.badssl.com/
		return modelx.FailureSSLInvalidCertificate
	}

	s := err.Error()
	if strings.HasSuffix(s, "operation was canceled") {
		return modelx.FailureInterrupted
	}
	if strings.HasSuffix(s, "EOF") {
		return modelx.FailureEOFError
	}
	if strings.HasSuffix(s, "connection refused") {
		return modelx.FailureConnectionRefused
	}
	if strings.HasSuffix(s, "connection reset by peer") {
		return modelx.FailureConnectionReset
	}
	if strings.HasSuffix(s, "context deadline exceeded") {
		return modelx.FailureGenericTimeoutError
	}
	if strings.HasSuffix(s, "transaction is timed out") {
		return modelx.FailureGenericTimeoutError
	}
	if strings.HasSuffix(s, "i/o timeout") {
		return modelx.FailureGenericTimeoutError
	}
	if strings.HasSuffix(s, "TLS handshake timeout") {
		return modelx.FailureGenericTimeoutError
	}
	if strings.HasSuffix(s, "no such host") {
		// This is dns_lookup_error in MK but such error is used as a
		// generic "hey, the lookup failed" error. Instead, this error
		// that we return here is significantly more specific.
		return modelx.FailureDNSNXDOMAINError
	}

	formatted := fmt.Sprintf("unknown_failure: %s", s)
	return Scrub(formatted) // scrub IP addresses in the error
}

func toOperationString(err error, operation string) string {
	var errwrapper *modelx.ErrWrapper
	if errors.As(err, &errwrapper) {
		// Basically, as explained in modelx.ErrWrapper docs, let's
		// keep the child major operation, if any.
		if errwrapper.Operation == modelx.ConnectOperation {
			return errwrapper.Operation
		}
		if errwrapper.Operation == modelx.HTTPRoundTripOperation {
			return errwrapper.Operation
		}
		if errwrapper.Operation == modelx.ResolveOperation {
			return errwrapper.Operation
		}
		if errwrapper.Operation == modelx.TLSHandshakeOperation {
			return errwrapper.Operation
		}
		// FALLTHROUGH
	}
	return operation
}
