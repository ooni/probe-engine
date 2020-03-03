// Package transactioner contains the transaction assigning round tripper
package transactioner

import (
	"net/http"

	"github.com/ooni/probe-engine/netx/internal/transactionid"
)

// Transport performs single HTTP transactions.
type Transport struct {
	roundTripper http.RoundTripper
}

// New creates a new Transport.
func New(roundTripper http.RoundTripper) *Transport {
	return &Transport{
		roundTripper: roundTripper,
	}
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTripper.RoundTrip(req.WithContext(
		transactionid.WithTransactionID(req.Context()),
	))
}

// CloseIdleConnections closes the idle connections.
func (t *Transport) CloseIdleConnections() {
	// Adapted from net/http code
	type closeIdler interface {
		CloseIdleConnections()
	}
	if tr, ok := t.roundTripper.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}
