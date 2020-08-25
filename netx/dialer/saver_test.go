package dialer_test

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestUnitSaverDialerFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	dlr := dialer.SaverDialer{
		Dialer: dialer.FakeDialer{
			Err: expected,
		},
		Saver: saver,
	}
	conn, err := dlr.DialContext(context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, expected) {
		t.Fatal("expected another error here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
	ev := saver.Read()
	if len(ev) != 1 {
		t.Fatal("expected a single event here")
	}
	if ev[0].Address != "www.google.com:443" {
		t.Fatal("unexpected Address")
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
	if ev[0].Proto != "tcp" {
		t.Fatal("unexpected Proto")
	}
	if !ev[0].Time.Before(time.Now()) {
		t.Fatal("unexpected Time")
	}
}

func TestUnitSaverConnDialerFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	dlr := dialer.SaverConnDialer{
		Dialer: dialer.FakeDialer{
			Err: expected,
		},
		Saver: saver,
	}
	conn, err := dlr.DialContext(context.Background(), "tcp", "www.google.com:443")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
}

func TestIntegrationSaverTLSHandshakerSuccessWithReadWrite(t *testing.T) {
	// This is the most common use case for collecting reads, writes
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	nextprotos := []string{"h2"}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Config: &tls.Config{NextProtos: nextprotos},
		Dialer: dialer.SaverConnDialer{
			Dialer: new(net.Dialer),
			Saver:  saver,
		},
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	// Implementation note: we don't close the connection here because it is
	// very handy to have the last event being the end of the handshake
	_, err := tlsdlr.DialTLSContext(context.Background(), "tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	ev := saver.Read()
	if len(ev) < 4 {
		// it's a bit tricky to be sure about the right number of
		// events because network conditions may influence that
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
	if ev[last].Duration <= 0 {
		t.Fatal("unexpected Duration")
	}
	if ev[last].Err != nil {
		t.Fatal("unexpected Err")
	}
	if ev[last].Name != "tls_handshake_done" {
		t.Fatal("unexpected Name")
	}
	if ev[last].TLSCipherSuite == "" {
		t.Fatal("unexpected TLSCipherSuite")
	}
	if ev[last].TLSNegotiatedProto != "h2" {
		t.Fatal("unexpected TLSNegotiatedProto")
	}
	if !reflect.DeepEqual(ev[last].TLSNextProtos, nextprotos) {
		t.Fatal("unexpected TLSNextProtos")
	}
	if ev[last].TLSPeerCerts == nil {
		t.Fatal("unexpected TLSPeerCerts")
	}
	if ev[last].TLSServerName != "www.google.com" {
		t.Fatal("unexpected TLSServerName")
	}
	if ev[last].TLSVersion == "" {
		t.Fatal("unexpected TLSVersion")
	}
	if ev[last].Time.Before(ev[last-1].Time) {
		t.Fatal("unexpected Time")
	}
}

func TestIntegrationSaverTLSHandshakerSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	nextprotos := []string{"h2"}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Config: &tls.Config{NextProtos: nextprotos},
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	conn, err := tlsdlr.DialTLSContext(context.Background(), "tcp", "www.google.com:443")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
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
		t.Fatal("unexpected Err")
	}
	if ev[1].Name != "tls_handshake_done" {
		t.Fatal("unexpected Name")
	}
	if ev[1].TLSCipherSuite == "" {
		t.Fatal("unexpected TLSCipherSuite")
	}
	if ev[1].TLSNegotiatedProto != "h2" {
		t.Fatal("unexpected TLSNegotiatedProto")
	}
	if !reflect.DeepEqual(ev[1].TLSNextProtos, nextprotos) {
		t.Fatal("unexpected TLSNextProtos")
	}
	if ev[1].TLSPeerCerts == nil {
		t.Fatal("unexpected TLSPeerCerts")
	}
	if ev[1].TLSServerName != "www.google.com" {
		t.Fatal("unexpected TLSServerName")
	}
	if ev[1].TLSVersion == "" {
		t.Fatal("unexpected TLSVersion")
	}
	if ev[1].Time.Before(ev[0].Time) {
		t.Fatal("unexpected Time")
	}
}

func TestIntegrationSaverTLSHandshakerHostnameError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	conn, err := tlsdlr.DialTLSContext(
		context.Background(), "tcp", "wrong.host.badssl.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
	for _, ev := range saver.Read() {
		if ev.Name != "tls_handshake_done" {
			continue
		}
		if ev.NoTLSVerify == true {
			t.Fatal("expected NoTLSVerify to be false")
		}
		if len(ev.TLSPeerCerts) < 1 {
			t.Fatal("expected at least a certificate here")
		}
	}
}

func TestIntegrationSaverTLSHandshakerInvalidCertError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	conn, err := tlsdlr.DialTLSContext(
		context.Background(), "tcp", "expired.badssl.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
	for _, ev := range saver.Read() {
		if ev.Name != "tls_handshake_done" {
			continue
		}
		if ev.NoTLSVerify == true {
			t.Fatal("expected NoTLSVerify to be false")
		}
		if len(ev.TLSPeerCerts) < 1 {
			t.Fatal("expected at least a certificate here")
		}
	}
}

func TestIntegrationSaverTLSHandshakerAuthorityError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	conn, err := tlsdlr.DialTLSContext(
		context.Background(), "tcp", "self-signed.badssl.com:443")
	if err == nil {
		t.Fatal("expected an error here")
	}
	if conn != nil {
		t.Fatal("expected nil conn here")
	}
	for _, ev := range saver.Read() {
		if ev.Name != "tls_handshake_done" {
			continue
		}
		if ev.NoTLSVerify == true {
			t.Fatal("expected NoTLSVerify to be false")
		}
		if len(ev.TLSPeerCerts) < 1 {
			t.Fatal("expected at least a certificate here")
		}
	}
}

func TestIntegrationSaverTLSHandshakerNoTLSVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	tlsdlr := dialer.TLSDialer{
		Config: &tls.Config{InsecureSkipVerify: true},
		Dialer: new(net.Dialer),
		TLSHandshaker: dialer.SaverTLSHandshaker{
			TLSHandshaker: dialer.SystemTLSHandshaker{},
			Saver:         saver,
		},
	}
	conn, err := tlsdlr.DialTLSContext(
		context.Background(), "tcp", "self-signed.badssl.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatal("expected non-nil conn here")
	}
	conn.Close()
	for _, ev := range saver.Read() {
		if ev.Name != "tls_handshake_done" {
			continue
		}
		if ev.NoTLSVerify != true {
			t.Fatal("expected NoTLSVerify to be true")
		}
		if len(ev.TLSPeerCerts) < 1 {
			t.Fatal("expected at least a certificate here")
		}
	}
}
