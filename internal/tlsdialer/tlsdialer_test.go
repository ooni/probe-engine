package tlsdialer_test

import (
	"context"
	"net"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/tlsdialer"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var h tlsdialer.Handshaker
	h = tlsdialer.StdlibHandshaker{}
	h = tlsdialer.ErrWrapper{Handshaker: h}
	saver := &tlsdialer.EventsSaver{Handshaker: h}
	h = saver
	h = tlsdialer.LoggingHandshaker{Handshaker: h, Logger: log.Log}
	var d tlsdialer.Dialer = tlsdialer.StdlibDialer{
		CleartextDialer: &net.Dialer{},
		Handshaker:      h,
	}
	conn, err := d.DialTLSContext(context.Background(), "tcp", "www.facebook.com:443")
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range saver.ReadEvents() {
		t.Logf("%+v", ev)
	}
	conn.Close()
}
