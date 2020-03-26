package dialer_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/dialer"
	"github.com/ooni/probe-engine/internal/httptransport"
	"github.com/ooni/probe-engine/internal/tlsdialer"
)

func TestIntegrationJustDialer(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var d dialer.Dialer
	d = dialer.Base()
	d = dialer.ErrWrapper{Dialer: d}
	saver := &dialer.EventsSaver{Dialer: d}
	d = saver
	d = dialer.LoggingDialer{Dialer: d, Logger: log.Log}
	d = dialer.ResolvingDialer{Connector: d, Resolver: net.DefaultResolver}
	d = dialer.LoggingDialer{Dialer: d, Logger: log.Log}
	conn, err := d.DialContext(context.Background(), "tcp", "www.facebook.com:80")
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range saver.ReadEvents() {
		t.Logf("%+v", ev)
	}
	conn.Close()
}

func TestIntegrationDialerService(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	svc := dialer.NewService()
	dial := dialer.LoggingDialer{
		Dialer: svc,
		Logger: log.Log,
	}
	tlsdial := tlsdialer.StdlibDialer{
		CleartextDialer: dial,
		Handshaker: tlsdialer.LoggingHandshaker{
			Handshaker: tlsdialer.StdlibHandshaker{},
			Logger:     log.Log,
		},
	}
	txp := httptransport.NewBase(dial, tlsdial)
	client := &http.Client{Transport: txp}
	ctx, cancel := context.WithCancel(context.Background())
	firsterr := errors.New("first error")
	if err := svc.Start(ctx, dialer.Mockable{Err: firsterr}); err != nil {
		t.Fatal(err)
	}
	_, err := client.Get("http://www.google.com")
	if !errors.Is(err, firsterr) {
		t.Fatal("not the error we expected")
	}
	cancel()
	ctx, cancel = context.WithCancel(context.Background())
	seconderr := errors.New("second error")
	if err := svc.Start(ctx, dialer.Mockable{Err: seconderr}); err != nil {
		t.Fatal(err)
	}
	_, err = client.Get("http://www.kernel.org")
	if !errors.Is(err, seconderr) {
		t.Fatal("not the error we expected")
	}
	cancel()
}

func TestUnitDialerServiceUniqueOwnership(t *testing.T) {
	svc := dialer.NewService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := svc.Start(ctx, dialer.Mockable{}); err != nil {
		t.Fatal(err)
	}
	if err := svc.Start(ctx, dialer.Mockable{}); !errors.Is(err, dialer.ErrBusy) {
		t.Fatal("not the error we expected")
	}
}
