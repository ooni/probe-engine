// Package urlgetter implements a nettest that fetches a URL. This is not
// an official OONI nettest, but rather is a probe-engine specific internal
// nettest that can be useful to do research.
//
// We manage the following URLs:
//
// 1. https://domain/path and http://domain/path trigger HTTP GETs
//
// 2. dnslookup://domain triggers a DNS lookup of domain
//
// 3. tlshandshake://domain:port triggers a TLS handshake connecting
// to the specified domain and port
//
// 4. other://domain:port triggers a TCP connect
//
// In all cases, the options specified inside of Config apply:
//
// 1. ResolverURL is used to configure a resolver. As a special case
// we recognize and properly handle `doh://google` and `doh://cloudflare`
//
// 2. TLSServerName allows to override the SNI
//
// 3. Tunnel allows to run the test inside a tunnel (e.g. psiphon)
//
// Because this is an experimental test, behaviour may change.
package urlgetter

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/experiment/httpheader"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	testName    = "urlgetter"
	testVersion = "0.0.2"
)

// Config contains the experiment's configuration.
type Config struct {
	ResolverURL   string `ooni:"URL describing a resolver"`
	TLSServerName string `ooni:"Force using a specific server name"`
	Tunnel        string `ooni:"Run experiment over a tunnel, e.g. psiphon"`
}

// TestKeys contains the experiment's result.
type TestKeys struct {
	Agent         string                   `json:"agent"`
	BootstrapTime float64                  `json:"bootstrap_time,omitempty"`
	NetworkEvents []archival.NetworkEvent  `json:"network_events"`
	Queries       []archival.DNSQueryEntry `json:"queries"`
	Requests      []archival.RequestEntry  `json:"requests"`
	SOCKSProxy    string                   `json:"socksproxy,omitempty"`
	TLSHandshakes []archival.TLSHandshake  `json:"tls_handshakes"`
	Tunnel        string                   `json:"tunnel,omitempty"`
}

// GetterConfig contains the configuration of the getter
type GetterConfig struct {
	Config
	Session model.ExperimentSession
	Target  string
}

// Do performs the action described in config and returns the TestKeys
// along with the error that occurred, if any.
func Do(ctx context.Context, config GetterConfig) (tk TestKeys, err error) {
	tk = TestKeys{Agent: "redirect", Tunnel: config.Tunnel}
	targetURL, err := url.Parse(config.Target)
	if err != nil {
		return
	}
	saver := new(trace.Saver)
	httpConfig := httptransport.Config{
		ContextByteCounting: true,
		DialSaver:           saver,
		HTTPSaver:           saver,
		Logger:              config.Session.Logger(),
		ReadWriteSaver:      saver,
		ResolveSaver:        saver,
		TLSSaver:            saver,
	}
	// configure resolver
	resolverURL, err := url.Parse(config.ResolverURL)
	if err != nil {
		return
	}
	switch resolverURL.Scheme {
	case "system":
	case "doh":
		if resolverURL.Host == "google" {
			config.ResolverURL = "https://dns.google/dns-query"
		} else if resolverURL.Host == "cloudflare" {
			config.ResolverURL = "https://cloudflare-dns.com/dns-query"
		} else {
			return tk, errors.New("unsupported doh://domain shortcut")
		}
		fallthrough // so we can manage this as an HTTPS URL
	case "https":
		dohClient := &http.Client{Transport: httptransport.New(httpConfig)}
		defer dohClient.CloseIdleConnections()
		httpConfig.BaseResolver = resolver.NewSerialResolver(
			resolver.SaverDNSTransport{
				RoundTripper: resolver.NewDNSOverHTTPS(
					dohClient, config.ResolverURL,
				),
				Saver: saver,
			},
		)
	default:
		err = errors.New("unsupported resolver scheme")
		return
	}
	// configure TLS
	if config.TLSServerName != "" {
		httpConfig.TLSConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			ServerName: config.TLSServerName,
		}
	}
	// configure tunnel
	if err = config.Session.MaybeStartTunnel(ctx, config.Tunnel); err != nil {
		return
	}
	tk.BootstrapTime = config.Session.TunnelBootstrapTime().Seconds()
	if url := config.Session.ProxyURL(); url != nil {
		tk.SOCKSProxy = url.Host
	}
	httpConfig.ProxyURL = config.Session.ProxyURL()
	// arrange for writing the results
	begin := time.Now()
	defer func() {
		events := saver.Read()
		tk.Queries = append(
			tk.Queries, archival.NewDNSQueriesList(begin, events)...,
		)
		tk.NetworkEvents = append(
			tk.NetworkEvents, archival.NewNetworkEventsList(begin, events)...,
		)
		tk.Requests = append(
			tk.Requests, archival.NewRequestList(begin, events)...,
		)
		tk.TLSHandshakes = append(
			tk.TLSHandshakes, archival.NewTLSHandshakesList(begin, events)...,
		)
	}()
	// perform the requested action
	switch targetURL.Scheme {
	case "http", "https":
		var (
			req  *http.Request
			resp *http.Response
		)
		req, err = http.NewRequest("GET", config.Target, nil)
		if err != nil {
			return
		}
		req = req.WithContext(ctx)
		req.Header.Set("Accept", httpheader.RandomAccept())
		req.Header.Set("Accept-Language", httpheader.RandomAcceptLanguage())
		req.Header.Set("User-Agent", httpheader.RandomUserAgent())
		httpClient := &http.Client{Transport: httptransport.New(httpConfig)}
		defer httpClient.CloseIdleConnections()
		resp, err = httpClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
			return
		}
	case "dnslookup":
		resolver := httptransport.NewResolver(httpConfig)
		if _, err = resolver.LookupHost(ctx, targetURL.Hostname()); err != nil {
			return
		}
	case "tlshandshake":
		var conn net.Conn
		tlsDialer := httptransport.NewTLSDialer(httpConfig)
		conn, err = tlsDialer.DialTLSContext(ctx, "tcp", targetURL.Host)
		if err != nil {
			return
		}
		conn.Close()
	default:
		var conn net.Conn
		dialer := httptransport.NewDialer(httpConfig)
		conn, err = dialer.DialContext(ctx, "tcp", targetURL.Host)
		if err != nil {
			return
		}
		conn.Close()
	}
	return
}

func registerExtensions(m *model.Measurement) {
	archival.ExtHTTP.AddTo(m)
	archival.ExtDNS.AddTo(m)
	archival.ExtNetevents.AddTo(m)
	archival.ExtTLSHandshake.AddTo(m)
}

type measurer struct {
	config Config
}

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	config := GetterConfig{
		Config:  m.config,
		Session: sess,
		Target:  string(measurement.Input),
	}
	tk, err := Do(ctx, config)
	measurement.TestKeys = tk
	return err
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
