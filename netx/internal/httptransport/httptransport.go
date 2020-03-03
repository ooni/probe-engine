// Package httptransport contains HTTP transport extensions. Here we
// define a http.Transport that emits events.
package httptransport

import (
	"net/http"

	"github.com/ooni/probe-engine/netx/internal/httptransport/bodytracer"
	"github.com/ooni/probe-engine/netx/internal/httptransport/tracetripper"
	"github.com/ooni/probe-engine/netx/internal/httptransport/transactioner"
)

// Transport performs single HTTP transactions and emits
// measurement events as they happen.
type Transport struct {
	roundTripper http.RoundTripper
}

// New creates a new Transport.
func New(roundTripper http.RoundTripper) *Transport {
	return &Transport{
		roundTripper: transactioner.New(bodytracer.New(
			tracetripper.New(roundTripper))),
	}
}

// RoundTrip executes a single HTTP transaction, returning
// a Response for the provided Request.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// Make sure we're not sending Go's default User-Agent
	// if the user has configured no user agent
	if req.Header.Get("User-Agent") == "" {
		req.Header["User-Agent"] = nil
	}
	return t.roundTripper.RoundTrip(req)
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
