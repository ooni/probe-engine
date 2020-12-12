package dialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

type MockQUICDialer struct {
	Dialer dialer.QUICContextDialer
	Sess   quic.EarlySession
	Err    error
}

func (d MockQUICDialer) DialContext(ctx context.Context, network, addr string, host string, tlsCfg *tls.Config, cfg *quic.Config) (quic.EarlySession, error) {
	if d.Dialer != nil {
		d.Dialer.DialContext(ctx, network, addr, host, tlsCfg, cfg)
	}
	return d.Sess, d.Err
}

func TestQUICSaverDialerFailure(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	dlr := dialer.QUICSaverDialer{
		QUICContextDialer: MockQUICDialer{
			Err: expected,
		},
		Saver: saver,
	}
	sess, err := dlr.DialContext(context.Background(), "udp", "", "www.google.com:443", tlsConf, &quic.Config{})
	if !errors.Is(err, expected) {
		t.Fatal("expected another error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
	ev := saver.Read()
	if len(ev) != 1 {
		t.Fatal("expected a single event here")
	}
	if ev[0].Address != "www.google.com:443" {
		t.Fatal("unexpected Address", ev[0].Address)
	}
	if ev[0].Duration <= 0 {
		t.Fatal("unexpected Duration")
	}
	if !errors.Is(ev[0].Err, expected) {
		t.Fatal("unexpected Err")
	}
	if ev[0].Name != errorx.ConnectOperation {
		t.Fatal("unexpected Name")
	}
	if ev[0].Proto != "udp" {
		t.Fatal("unexpected Proto")
	}
	if !ev[0].Time.Before(time.Now()) {
		t.Fatal("unexpected Time")
	}
}

func TestQUICSaverConnDialSuccess(t *testing.T) {
	tlsConf := &tls.Config{
		NextProtos: []string{"h3-29"},
	}
	saver := &trace.Saver{}
	systemdialer := dialer.QUICSystemDialer{Saver: saver}

	sess, err := systemdialer.DialContext(context.Background(), "udp", "216.58.212.164:443", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if sess == nil {
		t.Fatal("unexpected nil session")
	}
	ev := saver.Read()
	if len(ev) < 4 {
		// it's a bit tricky to be sure about the right number of
		// events because network conditions may influence that
		t.Fatal("unexpected number of events")
	}

	last := len(ev) - 1
	for idx := 1; idx < last; idx++ {
		if ev[idx].Data == nil {
			t.Fatal("unexpected Data")
		}
		if ev[idx].Duration <= 0 {
			t.Fatal("unexpected Duration")
		}
		if ev[idx].Err != nil {
			t.Fatal("unexpected Err")
		}
		if ev[idx].NumBytes <= 0 {
			t.Fatal("unexpected NumBytes")
		}
		switch ev[idx].Name {
		case errorx.ReadOperation, errorx.WriteOperation:
		default:
			t.Fatal("unexpected Name")
		}
		if ev[idx].Time.Before(ev[idx-1].Time) {
			t.Fatal("unexpected Time")
		}
	}
}

func TestQUICHandshakeSaverSuccess(t *testing.T) {
	nextprotos := []string{"h3-29"}
	servername := "www.google.com"
	tlsConf := &tls.Config{
		NextProtos: nextprotos,
		ServerName: servername,
	}
	saver := &trace.Saver{}
	dlr := dialer.QUICHandshakeSaver{
		Dialer: dialer.QUICSystemDialer{},
		Saver:  saver,
	}

	sess, err := dlr.DialContext(context.Background(), "udp", "216.58.212.164:443", "www.google.com:443", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if sess == nil {
		t.Fatal("unexpected nil sess")
	}
	ev := saver.Read()
	if len(ev) != 2 {
		t.Fatal("unexpected number of events")
	}
	if ev[0].Name != "tls_handshake_start" {
		t.Fatal("unexpected Name")
	}
	if ev[0].TLSServerName != "www.google.com" {
		t.Fatal("unexpected TLSServerName")
	}
	if !reflect.DeepEqual(ev[0].TLSNextProtos, nextprotos) {
		t.Fatal("unexpected TLSNextProtos")
	}
	if ev[0].Time.After(time.Now()) {
		t.Fatal("unexpected Time")
	}
	if ev[1].Duration <= 0 {
		t.Fatal("unexpected Duration")
	}
	if ev[1].Err != nil {
		t.Fatal("unexpected Err", ev[1].Err)
	}
	if ev[1].Name != "tls_handshake_done" {
		t.Fatal("unexpected Name")
	}
	if !reflect.DeepEqual(ev[1].TLSNextProtos, nextprotos) {
		t.Fatal("unexpected TLSNextProtos")
	}
	if ev[1].TLSServerName != "www.google.com" {
		t.Fatal("unexpected TLSServerName")
	}
	if ev[1].Time.Before(ev[0].Time) {
		t.Fatal("unexpected Time")
	}
}

func TestQUICHandshakeSaverHostNameError(t *testing.T) {
	nextprotos := []string{"h3-29"}
	servername := "wrong.host.badssl.com"
	tlsConf := &tls.Config{
		NextProtos: nextprotos,
		ServerName: servername,
	}
	saver := &trace.Saver{}
	dlr := dialer.QUICHandshakeSaver{
		Dialer: dialer.QUICSystemDialer{},
		Saver:  saver,
	}

	sess, err := dlr.DialContext(context.Background(), "udp", "216.58.212.164:443", "www.google.com:443", tlsConf, &quic.Config{})
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil sess here")
	}
	for _, ev := range saver.Read() {
		if ev.Name != "tls_handshake_done" {
			continue
		}
		if ev.NoTLSVerify == true {
			t.Fatal("expected NoTLSVerify to be false")
		}
		if !strings.Contains(ev.Err.Error(), "certificate is valid for www.google.com, not "+servername) {
			t.Fatal("unexpected error type")
		}
	}
}
