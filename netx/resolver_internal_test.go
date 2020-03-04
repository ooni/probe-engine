package netx

import (
	"net/http"
	"time"

	"github.com/ooni/probe-engine/netx/modelx"
)

// NewHTTPClientForDoH exposes the factory function we use internally for
// creating DoH optimised clients to unit tests.
func NewHTTPClientForDoH(beginning time.Time, handler modelx.Handler) *http.Client {
	return newHTTPClientForDoH(beginning, handler)
}
