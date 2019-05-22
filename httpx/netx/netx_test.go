package netx_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-engine/httpx/netx"
)

func TestDialContext(t *testing.T) {
	conn, err := (&netx.RetryingDialer{}).DialContext(
		context.Background(), "tcp", "www.google.com:80",
	)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}
