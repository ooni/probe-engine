package dialer_test

import (
	"context"
	"crypto/tls"
	"io"
	"testing"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/legacy/netx/dialid"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/errorx"
)

func TestQUICErrorWrapperFailure(t *testing.T) {
	ctx := dialid.WithDialID(context.Background())
	d := dialer.QUICErrorWrapperDialer{Dialer: MockQUICDialer{Sess: nil, Err: io.EOF}}
	sess, err := d.DialContext(ctx, "udp", "", "www.google.com:443", &tls.Config{}, &quic.Config{})
	if sess != nil {
		t.Fatal("expected a nil sess here")
	}
	errorWrapperCheckErr(t, err, errorx.QUICHandshakeOperation)
}

func TestQUICSystemDialerSuccess(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	systemdialer := dialer.QUICSystemDialer{}

	sess, err := systemdialer.DialContext(context.Background(), "udp", "216.58.212.164:443", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if sess == nil {
		t.Fatal("unexpected nil session")
	}
}
