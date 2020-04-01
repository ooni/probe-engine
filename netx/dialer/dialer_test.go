package dialer

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/ooni/probe-engine/netx/modelx"
)

func TestIntegrationDialerNew(t *testing.T) {
	var dialer modelx.Dialer = New(new(net.Resolver), new(net.Dialer))
	conn, err := dialer.Dial("tcp", "www.kernel.org:80")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("expected non-nil conn")
	}
	conn.Close()
}

func TestIntegrationDialerNewTLS(t *testing.T) {
	var dialer modelx.TLSDialer = NewTLS(new(net.Dialer), new(tls.Config))
	conn, err := dialer.DialTLS("tcp", "www.kernel.org:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("expected non-nil conn")
	}
	conn.Close()
}
