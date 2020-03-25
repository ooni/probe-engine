// Package measurer contains measurement code
package measurer

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ooni/probe-engine/internal/dialer"
	"github.com/ooni/probe-engine/internal/httptransport"
	"github.com/ooni/probe-engine/internal/resolver"
	"github.com/ooni/probe-engine/internal/tlsdialer"
)

// Logger is the logger interface assumed by this package
type Logger interface {
	Debugf(format string, v ...interface{})
}

// Config contains the settings for the measurer.
type Config struct {
	Logger Logger
}

// Results contains the measurement results.
type Results struct {
	Connects       []dialer.Events
	HTTPBodies     []httptransport.BodySnapshot
	HTTPRoundTrips []httptransport.Events
	Resolutions    []resolver.Events
	TLSHandshakes  []tlsdialer.Events
}

// Get performs a GET request and returns the measurement results. The error
// signals whether the request failed. You will get results in any case, since
// the purpose of this function is to measure anomalies :^).
func Get(url string, config Config) (Results, error) {
	g := newGetter(config)
	defer g.Close()
	resp, err := g.httpClient.Get(url)
	if err != nil {
		return g.results(), err
	}
	if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
		return g.results(), err
	}
	err = resp.Body.Close()
	return g.results(), err
}

type getter struct {
	config             Config
	dialerSaver        *dialer.EventsSaver
	httpBodySaver      *httptransport.SnapshotSaver
	httpClient         *http.Client
	httpTransportSaver *httptransport.EventsSaver
	resolverSaver      *resolver.EventsSaver
	tlsDialerSaver     *tlsdialer.EventsSaver
}

func (g *getter) Close() error {
	g.httpClient.CloseIdleConnections()
	return nil
}

func (g *getter) results() Results {
	return Results{
		Connects:       g.dialerSaver.ReadEvents(),
		HTTPBodies:     g.httpBodySaver.Snapshots(),
		HTTPRoundTrips: g.httpTransportSaver.ReadEvents(),
		Resolutions:    g.resolverSaver.ReadEvents(),
		TLSHandshakes:  g.tlsDialerSaver.ReadEvents(),
	}
}

func newGetter(config Config) *getter {
	g := &getter{config: config}
	res := g.newResolver()
	d := g.newDialer(res)
	td := g.newTLSDialer(d)
	g.newHTTPClient(d, td)
	return g
}

func (g *getter) newResolver() resolver.Resolver {
	var r resolver.Resolver = resolver.Base()
	r = resolver.ErrWrapper{Resolver: r}
	g.resolverSaver = &resolver.EventsSaver{Resolver: r}
	r = g.resolverSaver
	r = resolver.LoggingResolver{Resolver: r, Logger: g.config.Logger}
	return r
}

func (g *getter) newDialer(res resolver.Resolver) dialer.Dialer {
	var d dialer.Dialer = dialer.Base()
	d = dialer.ErrWrapper{Dialer: d}
	g.dialerSaver = &dialer.EventsSaver{Dialer: d}
	d = g.dialerSaver
	d = dialer.LoggingDialer{Dialer: d, Logger: g.config.Logger}
	d = dialer.ResolvingDialer{Connector: d, Resolver: res}
	d = dialer.LoggingDialer{Dialer: d, Logger: g.config.Logger}
	return d
}

func (g *getter) newTLSDialer(d dialer.Dialer) tlsdialer.Dialer {
	var h tlsdialer.Handshaker = tlsdialer.StdlibHandshaker{}
	h = tlsdialer.ErrWrapper{Handshaker: h}
	g.tlsDialerSaver = &tlsdialer.EventsSaver{Handshaker: h}
	h = g.tlsDialerSaver
	h = tlsdialer.LoggingHandshaker{Handshaker: h, Logger: g.config.Logger}
	return tlsdialer.StdlibDialer{
		CleartextDialer: d,
		Handshaker:      h,
	}
}

func (g *getter) newHTTPClient(d dialer.Dialer, td tlsdialer.Dialer) {
	var txp httptransport.Transport = httptransport.NewBase(d, td)
	txp = httptransport.ErrWrapper{Transport: txp}
	g.httpTransportSaver = &httptransport.EventsSaver{Transport: txp}
	txp = g.httpTransportSaver
	txp = httptransport.HeaderAdder{Transport: txp, UserAgent: "miniooni/0.1.0-dev"}
	g.httpBodySaver = &httptransport.SnapshotSaver{Transport: txp}
	txp = g.httpBodySaver
	txp = httptransport.Logging{Transport: txp, Logger: g.config.Logger}
	g.httpClient = &http.Client{Transport: txp}
}
