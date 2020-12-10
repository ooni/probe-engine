// Package dnscheck contains the DNS check experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-028-dnscheck.md.
package dnscheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	testName      = "dnscheck"
	testVersion   = "0.4.0"
	defaultDomain = "example.org"
)

// Config contains the experiment's configuration.
type Config struct {
	Domain        string `json:"domain" ooni:"domain to resolve using the specified resolver"`
	HTTP3Enabled  bool   `json:"http3_enabled" ooni:"use http3 instead of http/1.1 or http2"`
	HTTPHost      string `json:"http_host" ooni:"Force using specific HTTP Host header"`
	TLSServerName string `json:"tls_server_name" ooni:"force TLS to using a specific SNI in Client Hello"`
}

// TestKeys contains the results of the dnscheck experiment.
type TestKeys struct {
	Domain           string                        `json:"domain"`
	HTTP3Enabled     bool                          `json:"x_http3_enabled,omitempty"`
	HTTPHost         string                        `json:"x_http_host,omitempty"`
	TLSServerName    string                        `json:"x_tls_server_name,omitempty"`
	Bootstrap        *urlgetter.TestKeys           `json:"bootstrap"`
	BootstrapFailure *string                       `json:"bootstrap_failure"`
	Lookups          map[string]urlgetter.TestKeys `json:"lookups"`
}

// Measurer performs the measurement.
type Measurer struct {
	Config
}

// ExperimentName implements model.ExperimentSession.ExperimentName
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements model.ExperimentSession.ExperimentVersion
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

// The following errors may be returned by this experiment. Of course these
// errors are in addition to any other errors returned by the low level packages
// that are used by this experiment to implement its functionality.
var (
	ErrInputRequired        = errors.New("this experiment needs input")
	ErrInvalidURL           = errors.New("the input URL is invalid")
	ErrUnsupportedURLScheme = errors.New("unsupported URL scheme")
)

// Run implements model.ExperimentSession.Run
func (m Measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	// 1. fill the measurement with test keys
	tk := new(TestKeys)
	tk.Lookups = make(map[string]urlgetter.TestKeys)
	measurement.TestKeys = tk
	urlgetter.RegisterExtensions(measurement)

	// 2. make sure the runtime is bounded
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 3. select the domain to resolve or use default and, while there, also
	// ensure that we register all the other options we're using.
	domain := m.Config.Domain
	if domain == "" {
		domain = defaultDomain
	}
	tk.Domain = domain
	tk.HTTP3Enabled = m.Config.HTTP3Enabled
	tk.HTTPHost = m.Config.HTTPHost
	tk.TLSServerName = m.Config.TLSServerName

	// 4. parse the input URL describing the resolver to use
	input := string(measurement.Input)
	if input == "" {
		return ErrInputRequired
	}
	URL, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}
	switch URL.Scheme {
	case "https", "dot", "udp", "tcp":
		// all good
	default:
		return ErrUnsupportedURLScheme
	}

	// 5. possibly expand a domain to a list of IP addresses
	//
	// implementation note: because the resolver we constructed also deals
	// with IP addresses successfully, we just get back the IPs when we are
	// passing as input an IP address rather than a domain name.
	begin := time.Now()
	evsaver := new(trace.Saver)
	resolver := netx.NewResolver(netx.Config{
		BogonIsError: true,
		Logger:       sess.Logger(),
		ResolveSaver: evsaver,
	})
	addrs, err := resolver.LookupHost(ctx, URL.Hostname())
	queries := archival.NewDNSQueriesList(begin, evsaver.Read(), sess.ASNDatabasePath())
	tk.BootstrapFailure = archival.NewFailure(err)
	if len(queries) > 0 {
		// We get no queries in case we are resolving an IP address, since
		// the address resolver doesn't generate events
		tk.Bootstrap = &urlgetter.TestKeys{Queries: queries}
	}

	// 6. determine all the domain lookups we need to perform
	var inputs []urlgetter.MultiInput
	multi := urlgetter.Multi{Begin: begin, Session: sess}
	for _, addr := range addrs {
		inputs = append(inputs, urlgetter.MultiInput{
			Config: urlgetter.Config{
				DNSHTTPHost:      m.httpHost(URL.Host),            // use original host (and optional port)
				DNSTLSServerName: m.tlsServerName(URL.Hostname()), // just the domain/IP for SNI
				HTTP3Enabled:     m.Config.HTTP3Enabled,
				RejectDNSBogons:  true, // bogons are errors in this context
				ResolverURL:      makeResolverURL(URL, addr),
			},
			Target: fmt.Sprintf("dnslookup://%s", domain), // urlgetter wants a URL
		})
	}

	// 7. perform all the required resolutions
	for output := range Collect(ctx, multi, inputs, callbacks) {
		tk.Lookups[output.Input.Config.ResolverURL] = output.TestKeys
	}
	return nil
}

// httpHost returns the configured HTTP host, if set, otherwise
// it will return the host provide as argument.
func (m Measurer) httpHost(httpHost string) string {
	if m.Config.HTTPHost != "" {
		return m.Config.HTTPHost
	}
	return httpHost
}

// tlsServerName is like httpHost for the TLS server name.
func (m Measurer) tlsServerName(tlsServerName string) string {
	if m.Config.TLSServerName != "" {
		return m.Config.TLSServerName
	}
	return tlsServerName
}

// Collect prints on the output channel the result of running dnscheck
// on every provided input. It closes the output channel when done.
func Collect(ctx context.Context, multi urlgetter.Multi, inputs []urlgetter.MultiInput,
	callbacks model.ExperimentCallbacks) <-chan urlgetter.MultiOutput {
	outputch := make(chan urlgetter.MultiOutput)
	expect := len(inputs)
	inputch := multi.Run(ctx, inputs)
	go func() {
		var count int
		defer close(outputch)
		for count < expect {
			entry := <-inputch
			count++
			percentage := float64(count) / float64(expect)
			callbacks.OnProgress(percentage, fmt.Sprintf(
				"dnscheck: measure %s: %+v", entry.Input.Config.ResolverURL, entry.Err,
			))
			outputch <- entry
		}
	}()
	return outputch
}

// makeResolverURL rewrites the input URL to replace the domain in
// the input URL with the given addr. When the input URL already contains
// an addr, this operation will return the same URL.
func makeResolverURL(URL *url.URL, addr string) string {
	// 1. determine the hostname in the resulting URL
	hostname := URL.Hostname()
	if net.ParseIP(hostname) == nil {
		hostname = addr
	}
	// 2. adjust hostname if we also have a port
	if hasPort := URL.Port() != ""; hasPort {
		_, port, err := net.SplitHostPort(URL.Host)
		// We say this cannot fail because we already parsed the URL to validate
		// its scheme and hence the URL hostname should be well formed.
		runtimex.PanicOnError(err, "net.SplitHostPort should not fail here")
		hostname = net.JoinHostPort(hostname, port)
	} else if idx := strings.Index(addr, ":"); idx >= 0 {
		// Make sure an IPv6 address hostname without a port is properly
		// quoted to avoid breaking the URL parser down the line.
		hostname = "[" + addr + "]"
	}
	// 3. reassemble the URL
	return (&url.URL{
		Scheme:   URL.Scheme,
		Host:     hostname,
		Path:     URL.Path,
		RawQuery: URL.RawQuery,
	}).String()
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{Config: config}
}

// SummaryKeys contains summary keys for this experiment.
//
// Note that this structure is part of the ABI contract with probe-cli
// therefore we should be careful when changing it.
type SummaryKeys struct {
	IsAnomaly bool `json:"-"`
}

// GetSummaryKeys implements model.ExperimentMeasurer.GetSummaryKeys.
func (m Measurer) GetSummaryKeys(measurement *model.Measurement) (interface{}, error) {
	return SummaryKeys{IsAnomaly: false}, nil
}
