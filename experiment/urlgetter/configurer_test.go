package urlgetter_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestConfigurerNewConfigurationVanilla(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Logger: log.Log,
		Saver:  saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	defer configuration.CloseIdleConnections()
	if configuration.DNSOverHTTPClient != nil {
		t.Fatal("not the DNSOverHTTPClient we expected")
	}
	if configuration.HTTPConfig.BogonIsError != false {
		t.Fatal("not the BogonIsError we expected")
	}
	if configuration.HTTPConfig.ContextByteCounting != true {
		t.Fatal("not the ContextByteCounting we expected")
	}
	if configuration.HTTPConfig.DialSaver != saver {
		t.Fatal("not the DialSaver we expected")
	}
	if configuration.HTTPConfig.HTTPSaver != saver {
		t.Fatal("not the HTTPSaver we expected")
	}
	if configuration.HTTPConfig.Logger != log.Log {
		t.Fatal("not the Logger we expected")
	}
	if configuration.HTTPConfig.ReadWriteSaver != saver {
		t.Fatal("not the ReadWriteSaver we expected")
	}
	if configuration.HTTPConfig.ResolveSaver != saver {
		t.Fatal("not the ResolveSaver we expected")
	}
	if configuration.HTTPConfig.TLSSaver != saver {
		t.Fatal("not the TLSSaver we expected")
	}
	if configuration.HTTPConfig.BaseResolver != nil {
		t.Fatal("not the BaseResolver we expected")
	}
	if configuration.HTTPConfig.TLSConfig != nil {
		t.Fatal("not the TLSConfig we expected")
	}
	if configuration.HTTPConfig.ProxyURL != nil {
		t.Fatal("not the ProxyURL we expected")
	}
}

func TestConfigurerNewConfigurationResolverDNSOverHTTPSGoogle(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: "doh://google",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	defer configuration.CloseIdleConnections()
	if configuration.DNSOverHTTPClient == nil {
		t.Fatal("not the DNSOverHTTPClient we expected")
	}
	if configuration.HTTPConfig.BogonIsError != false {
		t.Fatal("not the BogonIsError we expected")
	}
	if configuration.HTTPConfig.ContextByteCounting != true {
		t.Fatal("not the ContextByteCounting we expected")
	}
	if configuration.HTTPConfig.DialSaver != saver {
		t.Fatal("not the DialSaver we expected")
	}
	if configuration.HTTPConfig.HTTPSaver != saver {
		t.Fatal("not the HTTPSaver we expected")
	}
	if configuration.HTTPConfig.Logger != log.Log {
		t.Fatal("not the Logger we expected")
	}
	if configuration.HTTPConfig.ReadWriteSaver != saver {
		t.Fatal("not the ReadWriteSaver we expected")
	}
	if configuration.HTTPConfig.ResolveSaver != saver {
		t.Fatal("not the ResolveSaver we expected")
	}
	if configuration.HTTPConfig.TLSSaver != saver {
		t.Fatal("not the TLSSaver we expected")
	}
	if configuration.HTTPConfig.BaseResolver == nil {
		t.Fatal("not the BaseResolver we expected")
	}
	sr, ok := configuration.HTTPConfig.BaseResolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	stxp, ok := sr.Txp.(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	dohtxp, ok := stxp.RoundTripper.(resolver.DNSOverHTTPS)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	if dohtxp.URL != "https://dns.google/dns-query" {
		t.Fatal("not the DoH URL we expected")
	}
	if configuration.HTTPConfig.TLSConfig != nil {
		t.Fatal("not the TLSConfig we expected")
	}
	if configuration.HTTPConfig.ProxyURL != nil {
		t.Fatal("not the ProxyURL we expected")
	}
}

func TestConfigurerNewConfigurationResolverDNSOverHTTPSCloudflare(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: "doh://cloudflare",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	defer configuration.CloseIdleConnections()
	if configuration.DNSOverHTTPClient == nil {
		t.Fatal("not the DNSOverHTTPClient we expected")
	}
	if configuration.HTTPConfig.BogonIsError != false {
		t.Fatal("not the BogonIsError we expected")
	}
	if configuration.HTTPConfig.ContextByteCounting != true {
		t.Fatal("not the ContextByteCounting we expected")
	}
	if configuration.HTTPConfig.DialSaver != saver {
		t.Fatal("not the DialSaver we expected")
	}
	if configuration.HTTPConfig.HTTPSaver != saver {
		t.Fatal("not the HTTPSaver we expected")
	}
	if configuration.HTTPConfig.Logger != log.Log {
		t.Fatal("not the Logger we expected")
	}
	if configuration.HTTPConfig.ReadWriteSaver != saver {
		t.Fatal("not the ReadWriteSaver we expected")
	}
	if configuration.HTTPConfig.ResolveSaver != saver {
		t.Fatal("not the ResolveSaver we expected")
	}
	if configuration.HTTPConfig.TLSSaver != saver {
		t.Fatal("not the TLSSaver we expected")
	}
	if configuration.HTTPConfig.BaseResolver == nil {
		t.Fatal("not the BaseResolver we expected")
	}
	sr, ok := configuration.HTTPConfig.BaseResolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	stxp, ok := sr.Txp.(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	dohtxp, ok := stxp.RoundTripper.(resolver.DNSOverHTTPS)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	if dohtxp.URL != "https://cloudflare-dns.com/dns-query" {
		t.Fatal("not the DoH URL we expected")
	}
	if configuration.HTTPConfig.TLSConfig != nil {
		t.Fatal("not the TLSConfig we expected")
	}
	if configuration.HTTPConfig.ProxyURL != nil {
		t.Fatal("not the ProxyURL we expected")
	}
}

func TestConfigurerNewConfigurationResolverUDP(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: "udp://8.8.8.8:53",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	defer configuration.CloseIdleConnections()
	if configuration.DNSOverHTTPClient != nil {
		t.Fatal("not the DNSOverHTTPClient we expected")
	}
	if configuration.HTTPConfig.BogonIsError != false {
		t.Fatal("not the BogonIsError we expected")
	}
	if configuration.HTTPConfig.ContextByteCounting != true {
		t.Fatal("not the ContextByteCounting we expected")
	}
	if configuration.HTTPConfig.DialSaver != saver {
		t.Fatal("not the DialSaver we expected")
	}
	if configuration.HTTPConfig.HTTPSaver != saver {
		t.Fatal("not the HTTPSaver we expected")
	}
	if configuration.HTTPConfig.Logger != log.Log {
		t.Fatal("not the Logger we expected")
	}
	if configuration.HTTPConfig.ReadWriteSaver != saver {
		t.Fatal("not the ReadWriteSaver we expected")
	}
	if configuration.HTTPConfig.ResolveSaver != saver {
		t.Fatal("not the ResolveSaver we expected")
	}
	if configuration.HTTPConfig.TLSSaver != saver {
		t.Fatal("not the TLSSaver we expected")
	}
	if configuration.HTTPConfig.BaseResolver == nil {
		t.Fatal("not the BaseResolver we expected")
	}
	sr, ok := configuration.HTTPConfig.BaseResolver.(resolver.SerialResolver)
	if !ok {
		t.Fatal("not the resolver we expected")
	}
	stxp, ok := sr.Txp.(resolver.SaverDNSTransport)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	udptxp, ok := stxp.RoundTripper.(resolver.DNSOverUDP)
	if !ok {
		t.Fatal("not the DNS transport we expected")
	}
	if udptxp.Address() != "8.8.8.8:53" {
		t.Fatal("not the DoH URL we expected")
	}
	if configuration.HTTPConfig.TLSConfig != nil {
		t.Fatal("not the TLSConfig we expected")
	}
	if configuration.HTTPConfig.ProxyURL != nil {
		t.Fatal("not the ProxyURL we expected")
	}
}

func TestConfigurerNewConfigurationResolverInvalidURL(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: "\t",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	_, err := configurer.NewConfiguration()
	if err == nil || !strings.HasSuffix(err.Error(), "invalid control character in URL") {
		t.Fatal("not the error we expected")
	}
}

func TestConfigurerNewConfigurationResolverInvalidURLScheme(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			ResolverURL: "antani://8.8.8.8:53",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	_, err := configurer.NewConfiguration()
	if err == nil || !strings.HasSuffix(err.Error(), "unsupported resolver scheme") {
		t.Fatal("not the error we expected")
	}
}

func TestConfigurerNewConfigurationTLSServerName(t *testing.T) {
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Config: urlgetter.Config{
			TLSServerName: "www.x.org",
		},
		Logger: log.Log,
		Saver:  saver,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if configuration.HTTPConfig.TLSConfig.ServerName != "www.x.org" {
		t.Fatal("invalid ServerName")
	}
	if len(configuration.HTTPConfig.TLSConfig.NextProtos) != 2 {
		t.Fatal("invalid len(NextProtos)")
	}
	if configuration.HTTPConfig.TLSConfig.NextProtos[0] != "h2" {
		t.Fatal("invalid NextProtos[0]")
	}
	if configuration.HTTPConfig.TLSConfig.NextProtos[1] != "http/1.1" {
		t.Fatal("invalid NextProtos[1]")
	}
}

func TestConfigurerNewConfigurationProxyURL(t *testing.T) {
	URL, _ := url.Parse("socks5://127.0.0.1:9050")
	saver := new(trace.Saver)
	configurer := urlgetter.Configurer{
		Logger:   log.Log,
		Saver:    saver,
		ProxyURL: URL,
	}
	configuration, err := configurer.NewConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if configuration.HTTPConfig.ProxyURL != URL {
		t.Fatal("invalid ProxyURL")
	}
}
