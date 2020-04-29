// Package urlgetter implements a nettest that fetches a URL
package urlgetter

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
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
	testVersion = "0.0.1"
)

// Config contains the experiment's configuration.
type Config struct {
	ResolverURL   string `ooni:"URL describing a resolver"`
	TLSServerName string `ooni:"Force using a specific server name"`
	Tunnel        string `ooni:"Run experiment over a tunnel, e.g. psiphon"`
}

// TestKeys contains the experiment's result.
type TestKeys struct {
	Agent         string                     `json:"agent"`
	BootstrapTime float64                    `json:"bootstrap_time,omitempty"`
	NetworkEvents archival.NetworkEventsList `json:"network_events"`
	Queries       archival.DNSQueriesList    `json:"queries"`
	Requests      archival.RequestList       `json:"requests"`
	SOCKSProxy    string                     `json:"socksproxy,omitempty"`
	TLSHandshakes archival.TLSHandshakesList `json:"tls_handshakes"`
	Tunnel        string                     `json:"tunnel,omitempty"`
}

func (tk *TestKeys) doget(
	clnt *http.Client, req *http.Request, saver *trace.Saver) error {
	saver.Write(trace.Event{Name: "http_transaction_start", Time: time.Now()})
	defer func() {
		saver.Write(trace.Event{Name: "http_transaction_done", Time: time.Now()})
	}()
	resp, err := clnt.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)
	return err
}

func (tk *TestKeys) get(
	ctx context.Context, measurement *model.Measurement,
	config httptransport.Config, saver *trace.Saver,
) error {
	clnt := &http.Client{Transport: httptransport.New(config)}
	defer clnt.CloseIdleConnections()
	req, err := http.NewRequest("GET", string(measurement.Input), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", httpheader.RandomAccept())
	req.Header.Set("Accept-Language", httpheader.RandomAcceptLanguage())
	req.Header.Set("User-Agent", httpheader.RandomUserAgent())
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
	return tk.doget(clnt, req.WithContext(ctx), saver)
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

// TODO(bassosimone): a measurer must declare whether it takes input since
// this is better than declaring that inside of experiment.go

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) maybeNewResolver(
	saver *trace.Saver, logger model.Logger) (resolver.Resolver, error) {
	var r resolver.Resolver
	resolverURL, err := url.Parse(m.config.ResolverURL)
	if err != nil {
		return nil, err
	}
	switch resolverURL.Scheme {
	case "https":
		// TODO(bassosimone): we are leaking connections on this client.
		httpClient := &http.Client{Transport: httptransport.New(httptransport.Config{
			ContextByteCounting: true,
			Logger:              logger,
			Saver:               saver,
		})}
		r = &resolver.CacheResolver{
			Resolver: resolver.SaverResolver{
				Resolver: resolver.LoggingResolver{
					Resolver: resolver.ErrorWrapperResolver{
						Resolver: resolver.NewSerialResolver(
							resolver.SaverDNSTransport{
								RoundTripper: resolver.NewDNSOverHTTPS(
									httpClient, m.config.ResolverURL,
								),
								Saver: saver,
							},
						),
					},
					Logger: logger,
				},
				Saver: saver,
			},
		}
	case "system":
	default:
		return nil, errors.New("unsupported resolver URL")
	}
	return r, nil
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	saver := new(trace.Saver)
	config := httptransport.Config{
		ContextByteCounting: true,
		Logger:              sess.Logger(),
		SaveReadWrite:       true,
		Saver:               saver,
	}
	reso, err := m.maybeNewResolver(saver, sess.Logger())
	if err != nil {
		return err
	}
	config.Resolver = reso
	if m.config.TLSServerName != "" {
		config.TLSConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			ServerName: m.config.TLSServerName,
		}
	}
	tk := new(TestKeys)
	measurement.TestKeys = tk
	tk.Agent = "redirect"
	tk.Tunnel = m.config.Tunnel
	if err := sess.MaybeStartTunnel(ctx, m.config.Tunnel); err != nil {
		return err
	}
	tk.BootstrapTime = sess.TunnelBootstrapTime().Seconds()
	if url := sess.ProxyURL(); url != nil {
		tk.SOCKSProxy = url.Host
	}
	config.ProxyURL = sess.ProxyURL()
	return tk.get(ctx, measurement, config, saver)
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
