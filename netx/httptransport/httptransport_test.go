package httptransport_test

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestNewResolverVanilla(t *testing.T) {
	r := httptransport.NewResolver(httptransport.Config{})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		BaseResolver: resolver.BogonResolver{
			// not initialized because it doesn't matter in this context
		},
	})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		BogonIsError: true,
	})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		Logger: log.Log,
	})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		ResolveSaver: saver,
	})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		CacheResolutions: true,
	})
	ar, ok := r.(resolver.AddressResolver)
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
	r := httptransport.NewResolver(httptransport.Config{
		DNSCache: map[string][]string{
			"dns.google.com": {"8.8.8.8"},
		},
	})
	ar, ok := r.(resolver.AddressResolver)
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
	d := httptransport.NewDialer(httptransport.Config{})
	pd, ok := d.(dialer.ProxyDialer)
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
	if _, ok := dnsd.Resolver.(resolver.AddressResolver); !ok {
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
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithResolver(t *testing.T) {
	d := httptransport.NewDialer(httptransport.Config{
		FullResolver: resolver.BogonResolver{
			// not initialized because it doesn't matter in this context
		},
	})
	pd, ok := d.(dialer.ProxyDialer)
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
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithLogger(t *testing.T) {
	d := httptransport.NewDialer(httptransport.Config{
		Logger: log.Log,
	})
	pd, ok := d.(dialer.ProxyDialer)
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
	if _, ok := dnsd.Resolver.(resolver.AddressResolver); !ok {
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
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithDialSaver(t *testing.T) {
	saver := new(trace.Saver)
	d := httptransport.NewDialer(httptransport.Config{
		DialSaver: saver,
	})
	pd, ok := d.(dialer.ProxyDialer)
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
	if _, ok := dnsd.Resolver.(resolver.AddressResolver); !ok {
		t.Fatal("not the resolver we expected")
	}
	sd, ok := dnsd.Dialer.(dialer.SaverDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if sd.Saver != saver {
		t.Fatal("not the logger we expected")
	}
	ewd, ok := sd.Dialer.(dialer.ErrorWrapperDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	td, ok := ewd.Dialer.(dialer.TimeoutDialer)
	if !ok {
		t.Fatal("not the dialer we expected")
	}
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithReadWriteSaver(t *testing.T) {
	saver := new(trace.Saver)
	d := httptransport.NewDialer(httptransport.Config{
		ReadWriteSaver: saver,
	})
	pd, ok := d.(dialer.ProxyDialer)
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
	if _, ok := dnsd.Resolver.(resolver.AddressResolver); !ok {
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
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewDialerWithContextByteCounting(t *testing.T) {
	d := httptransport.NewDialer(httptransport.Config{
		ContextByteCounting: true,
	})
	bcd, ok := d.(dialer.ByteCounterDialer)
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
	if _, ok := dnsd.Resolver.(resolver.AddressResolver); !ok {
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
	if _, ok := td.Dialer.(*net.Dialer); !ok {
		t.Fatal("not the dialer we expected")
	}
}

func TestNewTLSDialerVanilla(t *testing.T) {
	td := httptransport.NewTLSDialer(httptransport.Config{})
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
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	td := httptransport.NewTLSDialer(httptransport.Config{
		TLSConfig: new(tls.Config),
	})
	rtd, ok := td.(dialer.TLSDialer)
	if !ok {
		t.Fatal("not the TLSDialer we expected")
	}
	if len(rtd.Config.NextProtos) != 0 {
		t.Fatal("invalid len(config.NextProtos)")
	}
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	td := httptransport.NewTLSDialer(httptransport.Config{
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
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	td := httptransport.NewTLSDialer(httptransport.Config{
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
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	td := httptransport.NewTLSDialer(httptransport.Config{
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
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	td := httptransport.NewTLSDialer(httptransport.Config{
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
	if rtd.Dialer == nil {
		t.Fatal("invalid Dialer")
	}
	if _, ok := rtd.Dialer.(dialer.ProxyDialer); !ok {
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
	txp := httptransport.New(httptransport.Config{})
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
	dialer := httptransport.FakeDialer{Err: expected}
	txp := httptransport.New(httptransport.Config{
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
		Dialer:        httptransport.FakeDialer{Err: expected},
		TLSHandshaker: dialer.SystemTLSHandshaker{},
	}
	txp := httptransport.New(httptransport.Config{
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
	txp := httptransport.New(httptransport.Config{
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
	txp := httptransport.New(httptransport.Config{
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
	txp := httptransport.New(httptransport.Config{
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
