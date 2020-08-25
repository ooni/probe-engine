package netx

import (
	"net/http"
	"time"

	"github.com/ooni/probe-engine/legacy/netx/modelx"
)

func NewHTTPClientForDoH(beginning time.Time, handler modelx.Handler) *http.Client {
	return newHTTPClientForDoH(beginning, handler)
}

type ChainWrapperResolver = chainWrapperResolver
