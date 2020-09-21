package netx_test

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/selfcensor"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestNewResolverVanilla(t *testing.T) {
	r := netx.NewResolver(netx.Config{})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ewr, ok := ar.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	_, ok = ewr.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverSpecificResolver(t *testing.T) {
	r := netx.NewResolver(netx.Config{
		BaseResolver: resolver.BogonResolver{
			// not initialized because it doesn't matter in this context
		},
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ewr, ok := ar.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	_, ok = ewr.Resolver.(resolver.BogonResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverWithBogonFilter(t *testing.T) {
	r := netx.NewResolver(netx.Config{
		BogonIsError: true,
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ewr, ok := ar.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	br, ok := ewr.Resolver.(resolver.BogonResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	_, ok = br.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverWithLogging(t *testing.T) {
	r := netx.NewResolver(netx.Config{
		Logger: log.Log,
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	lr, ok := ar.Resolver.(resolver.LoggingResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if lr.Logger != log.Log {
		t.Fatal("not the logger we expected")
	}
	ewr, ok := lr.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	_, ok = ewr.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverWithSaver(t *testing.T) {
	saver := new(trace.Saver)
	r := netx.NewResolver(netx.Config{
		ResolveSaver: saver,
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	sr, ok := ar.Resolver.(resolver.SaverResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if sr.Saver != saver {
		t.Fatal("not the saver we expected")
	}
	ewr, ok := sr.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	_, ok = ewr.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverWithReadWriteCache(t *testing.T) {
	r := netx.NewResolver(netx.Config{
		CacheResolutions: true,
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ewr, ok := ar.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	cr, ok := ewr.Resolver.(*resolver.CacheResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if cr.ReadOnly != false {
		t.Fatal("expected readwrite cache here")
	}
	_, ok = cr.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewResolverWithPrefilledReadonlyCache(t *testing.T) {
	r := netx.NewResolver(netx.Config{
		DNSCache: map[string][]string{
			"dns.google.com": {"8.8.8.8"},
		},
	})
	ir, ok := r.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ar, ok := ir.Resolver.(resolver.AddressResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	ewr, ok := ar.Resolver.(resolver.ErrorWrapperResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	cr, ok := ewr.Resolver.(*resolver.CacheResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if cr.ReadOnly != true {
		t.Fatal("expected readonly cache here")
	}
	if cr.Get("dns.google.com")[0] != "8.8.8.8" {
		t.Fatal("cache not correctly prefilled")
	}
	_, ok = cr.Resolver.(resolver.SystemResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
}

func TestNewDialerVanilla(t *testing.T) {
	d := netx.NewDialer(netx.Config{})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := sd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	ir, ok := dnsd.Resolver.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := ir.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	ewd, ok := dnsd.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithResolver(t *testing.T) {
	d := netx.NewDialer(netx.Config{
		FullResolver: resolver.BogonResolver{
			// not initialized because it doesn't matter in this context
		},
	})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := sd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := dnsd.Resolver.(resolver.BogonResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	ewd, ok := dnsd.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithLogger(t *testing.T) {
	d := netx.NewDialer(netx.Config{
		Logger: log.Log,
	})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := sd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	ir, ok := dnsd.Resolver.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := ir.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	ld, ok := dnsd.Dialer.(dialer.LoggingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if ld.Logger != log.Log {
		t.Fatal("not the logger we expected")
	}
	ewd, ok := ld.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithDialSaver(t *testing.T) {
	saver := new(trace.Saver)
	d := netx.NewDialer(netx.Config{
		DialSaver: saver,
	})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := sd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	ir, ok := dnsd.Resolver.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := ir.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	sad, ok := dnsd.Dialer.(dialer.SaverDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if sad.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	ewd, ok := sad.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithReadWriteSaver(t *testing.T) {
	saver := new(trace.Saver)
	d := netx.NewDialer(netx.Config{
		ReadWriteSaver: saver,
	})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := sd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	ir, ok := dnsd.Resolver.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := ir.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	scd, ok := dnsd.Dialer.(dialer.SaverConnDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if scd.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	ewd, ok := scd.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithContextByteCounting(t *testing.T) {
	d := netx.NewDialer(netx.Config{
		ContextByteCounting: true,
	})
	sd, ok := d.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	bcd, ok := sd.Dialer.(dialer.ByteCounterDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	pd, ok := bcd.Dialer.(dialer.ProxyDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if pd.ProxyURL != nil {
		t.Fatal("not the proxy URL we expected")
	}
	dnsd, ok := pd.Dialer.(dialer.DNSDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if dnsd.Resolver == nil {
		t.Fatal("not the resolver we expected")
	}
	ir, ok := dnsd.Resolver.(resolver.IDNAResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := ir.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	ewd, ok := dnsd.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(selfcensor.SystemDialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewTLSDialerVanilla(t *testing.T) {
	td := netx.NewTLSDialer(netx.Config{})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 2 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.NextProtos[0] != "h2" || rtd.Config.NextProtos[1] != "http/1.1" {
		t.Fatal("invalid Config.NextProtos")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	ewth, ok := rtd.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewTLSDialerWithConfig(t *testing.T) {
	td := netx.NewTLSDialer(netx.Config{
		TLSConfig: new(tls.Config),
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 0 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	ewth, ok := rtd.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewTLSDialerWithLogging(t *testing.T) {
	td := netx.NewTLSDialer(netx.Config{
		Logger: log.Log,
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 2 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.NextProtos[0] != "h2" || rtd.Config.NextProtos[1] != "http/1.1" {
		t.Fatal("invalid Config.NextProtos")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	lth, ok := rtd.TLSHandshaker.(dialer.LoggingTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if lth.Logger != log.Log {
		t.Fatal("not the Logger we expected")
	}
	ewth, ok := lth.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewTLSDialerWithSaver(t *testing.T) {
	saver := new(trace.Saver)
	td := netx.NewTLSDialer(netx.Config{
		TLSSaver: saver,
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 2 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.NextProtos[0] != "h2" || rtd.Config.NextProtos[1] != "http/1.1" {
		t.Fatal("invalid Config.NextProtos")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	sth, ok := rtd.TLSHandshaker.(dialer.SaverTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if sth.Saver != saver {
		t.Fatal("not the Logger we expected")
	}
	ewth, ok := sth.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewTLSDialerWithNoTLSVerifyAndConfig(t *testing.T) {
	td := netx.NewTLSDialer(netx.Config{
		TLSConfig:   new(tls.Config),
		NoTLSVerify: true,
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 0 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.InsecureSkipVerify != true {
		t.Fatal("expected true InsecureSkipVerify")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	ewth, ok := rtd.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewTLSDialerWithNoTLSVerifyAndNoConfig(t *testing.T) {
	td := netx.NewTLSDialer(netx.Config{
		NoTLSVerify: true,
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 2 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Config.NextProtos[0] != "h2" || rtd.Config.NextProtos[1] != "http/1.1" {
		t.Fatal("invalid Config.NextProtos")
	}
	if rtd.Config.InsecureSkipVerify != true {
		t.Fatal("expected true InsecureSkipVerify")
	}
	if rtd.Config.RootCAs != netx.CertPool {
		t.Fatal("invalid Config.RootCAs")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	sd, ok := rtd.Dialer.(dialer.ShapingDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := sd.Dialer.(dialer.ProxyDialer); !ok {
		t.Fatal("not the Dialer we expected")
	}
	if rtd.TLSHandshaker == nil {
		t.Fatal("invalid TLSHandshaker")
	}
	ewth, ok := rtd.TLSHandshaker.(dialer.ErrorWrapperTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	tth, ok := ewth.TLSHandshaker.(dialer.TimeoutTLSHandshaker)
	if !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
	if _, ok := tth.TLSHandshaker.(dialer.SystemTLSHandshaker); !ok {
		t.Fatal("not the TLSHandshaker we expected")
	}
}

func TestNewVanilla(t *testing.T) {
	txp := netx.NewHTTPTransport(netx.Config{})
	uatxp, ok := txp.(httptransport.UserAgentTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if _, ok := uatxp.RoundTripper.(*http.Transport); !ok {
		t.Fatal("not the transport we expected")
	}
}

func TestNewWithDialer(t *testing.T) {
	expected := errors.New("mocked error")
	dialer := netx.FakeDialer{Err: expected}
	txp := netx.NewHTTPTransport(netx.Config{
		Dialer: dialer,
	})
	client := &http.Client{Transport: txp}
	resp, err := client.Get("http://www.google.com")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("not the response we expected")
	}
}

func TestNewWithTLSDialer(t *testing.T) {
	expected := errors.New("mocked error")
	tlsDialer := dialer.TLSDialer{
		Config:        new(tls.Config),
		Dialer:        netx.FakeDialer{Err: expected},
		TLSHandshaker: dialer.SystemTLSHandshaker{},
	}
	txp := netx.NewHTTPTransport(netx.Config{
		TLSDialer: tlsDialer,
	})
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com")
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("not the response we expected")
	}
}

func TestNewWithByteCounter(t *testing.T) {
	counter := bytecounter.New()
	txp := netx.NewHTTPTransport(netx.Config{
		ByteCounter: counter,
	})
	uatxp, ok := txp.(httptransport.UserAgentTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	bctxp, ok := uatxp.RoundTripper.(httptransport.ByteCountingTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if bctxp.Counter != counter {
		t.Fatal("not the byte counter we expected")
	}
	if _, ok := bctxp.RoundTripper.(*http.Transport); !ok {
		t.Fatal("not the transport we expected")
	}
}

func TestNewWithLogger(t *testing.T) {
	txp := netx.NewHTTPTransport(netx.Config{
		Logger: log.Log,
	})
	uatxp, ok := txp.(httptransport.UserAgentTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	ltxp, ok := uatxp.RoundTripper.(httptransport.LoggingTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if ltxp.Logger != log.Log {
		t.Fatal("not the logger we expected")
	}
	if _, ok := ltxp.RoundTripper.(*http.Transport); !ok {
		t.Fatal("not the transport we expected")
	}
}

func TestNewWithSaver(t *testing.T) {
	saver := new(trace.Saver)
	txp := netx.NewHTTPTransport(netx.Config{
		HTTPSaver: saver,
	})
	uatxp, ok := txp.(httptransport.UserAgentTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	stxptxp, ok := uatxp.RoundTripper.(httptransport.SaverTransactionHTTPTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if stxptxp.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	sptxp, ok := stxptxp.RoundTripper.(httptransport.SaverPerformanceHTTPTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if sptxp.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	sbtxp, ok := sptxp.RoundTripper.(httptransport.SaverBodyHTTPTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if sbtxp.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	smtxp, ok := sbtxp.RoundTripper.(httptransport.SaverMetadataHTTPTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if smtxp.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	if _, ok := smtxp.RoundTripper.(*http.Transport); !ok {
		t.Fatal("not the transport we expected")
	}
}

func TestNewDNSClientInvalidURL(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(netx.Config{}, "\t\t\t")
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
	if dnsclient.Resolver != nil {
		t.Fatal("expected nil resolver here")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientUnsupportedScheme(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(netx.Config{}, "antani:///")
	if err == nil || err.Error() != "unsupported resolver scheme" {
		t.Fatal("not the error we expected")
	}
	if dnsclient.Resolver != nil {
		t.Fatal("expected nil resolver here")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientSystemResolver(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "system:///")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := dnsclient.Resolver.(resolver.SystemResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientEmpty(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := dnsclient.Resolver.(resolver.SystemResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientPowerdnsDoH(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "doh://powerdns")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := r.Transport().(resolver.DNSOverHTTPS); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientGoogleDoH(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "doh://google")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := r.Transport().(resolver.DNSOverHTTPS); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientCloudflareDoH(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "doh://cloudflare")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := r.Transport().(resolver.DNSOverHTTPS); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientCloudflareDoHSaver(t *testing.T) {
	saver := new(trace.Saver)
	dnsclient, err := netx.NewDNSClient(
		netx.Config{ResolveSaver: saver}, "doh://cloudflare")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if _, ok := txp.RoundTripper.(resolver.DNSOverHTTPS); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientUDP(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "udp://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	if _, ok := r.Transport().(resolver.DNSOverUDP); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientUDPDNSSaver(t *testing.T) {
	saver := new(trace.Saver)
	dnsclient, err := netx.NewDNSClient(
		netx.Config{ResolveSaver: saver}, "udp://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if _, ok := txp.RoundTripper.(resolver.DNSOverUDP); !ok {
		t.Fatal("not the transport we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientTCP(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "tcp://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.DNSOverTCP)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if txp.Network() != "tcp" {
		t.Fatal("not the Network we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientTCPDNSSaver(t *testing.T) {
	saver := new(trace.Saver)
	dnsclient, err := netx.NewDNSClient(
		netx.Config{ResolveSaver: saver}, "tcp://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	dotcp, ok := txp.RoundTripper.(resolver.DNSOverTCP)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if dotcp.Network() != "tcp" {
		t.Fatal("not the Network we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientDoT(t *testing.T) {
	dnsclient, err := netx.NewDNSClient(
		netx.Config{}, "dot://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.DNSOverTCP)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if txp.Network() != "dot" {
		t.Fatal("not the Network we expected")
	}
	dnsclient.CloseIdleConnections()
}

func TestNewDNSClientDoTDNSSaver(t *testing.T) {
	saver := new(trace.Saver)
	dnsclient, err := netx.NewDNSClient(
		netx.Config{ResolveSaver: saver}, "dot://8.8.8.8:53")
	if err != nil {
		t.Fatal(err)
	}
	r, ok := dnsclient.Resolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	txp, ok := r.Transport().(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	dotls, ok := txp.RoundTripper.(resolver.DNSOverTCP)
	if !ok {
		t.Fatal("not the transport we expected")
	}
	if dotls.Network() != "dot" {
		t.Fatal("not the Network we expected")
	}
	dnsclient.CloseIdleConnections()
}
