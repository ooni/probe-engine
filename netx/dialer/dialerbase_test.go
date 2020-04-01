package dialer

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/handlers"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegrationBaseDialerSuccess(t *testing.T) {
	dialer := newBaseDialer()
	conn, err := dialer.Dial("tcp", "8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}

func TestIntegrationBaseDialerErrorNoConnect(t *testing.T) {
	dialer := newBaseDialer()
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp", "8.8.8.8:53")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if ctx.Err() == nil {
		t.Fatal("expected context to be expired here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

// see whether we implement the interface
func newBaseDialer() modelx.Dialer {
	return NewBaseDialer(
		time.Now(), handlers.NoHandler, new(net.Dialer), 17,
	)
}
