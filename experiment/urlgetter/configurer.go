package urlgetter

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/resolver"
	"github.com/ooni/probe-engine/netx/trace"
)

// The Configurer job is to construct a Configuration that can
// later be used by the measurer to perform measurements.
type Configurer struct {
	Config   Config
	Logger   model.Logger
	ProxyURL *url.URL
	Saver    *trace.Saver
}

// The Configuration is the configuration for running a measurement.
type Configuration struct {
	HTTPConfig        httptransport.Config
	DNSOverHTTPClient *http.Client
}

// CloseIdleConnections will close idle connections, if needed.
func (c Configuration) CloseIdleConnections() {
	if c.DNSOverHTTPClient != nil {
		c.DNSOverHTTPClient.CloseIdleConnections()
	}
}

// NewConfiguration builds a new measurement configuration.
func (c Configurer) NewConfiguration() (Configuration, error) {
	// set up defaults
	configuration := Configuration{
		HTTPConfig: httptransport.Config{
			BogonIsError:        c.Config.RejectDNSBogons,
			ContextByteCounting: true,
			DialSaver:           c.Saver,
			HTTPSaver:           c.Saver,
			Logger:              c.Logger,
			ReadWriteSaver:      c.Saver,
			ResolveSaver:        c.Saver,
			TLSSaver:            c.Saver,
		},
	}
	// configure the resolver
	switch c.Config.ResolverURL {
	case "doh://google":
		c.Config.ResolverURL = "https://dns.google/dns-query"
	case "doh://cloudflare":
		c.Config.ResolverURL = "https://cloudflare-dns.com/dns-query"
	case "":
		c.Config.ResolverURL = "system:///"
	}
	resolverURL, err := url.Parse(c.Config.ResolverURL)
	if err != nil {
		return configuration, err
	}
	switch resolverURL.Scheme {
	case "system":
	case "https":
		configuration.DNSOverHTTPClient = &http.Client{
			Transport: httptransport.New(configuration.HTTPConfig),
		}
		configuration.HTTPConfig.BaseResolver = resolver.NewSerialResolver(
			resolver.SaverDNSTransport{
				RoundTripper: resolver.NewDNSOverHTTPS(
					configuration.DNSOverHTTPClient, c.Config.ResolverURL,
				),
				Saver: c.Saver,
			},
		)
	case "udp":
		dialer := httptransport.NewDialer(configuration.HTTPConfig)
		configuration.HTTPConfig.BaseResolver = resolver.NewSerialResolver(
			resolver.SaverDNSTransport{
				RoundTripper: resolver.NewDNSOverUDP(
					dialer, resolverURL.Host,
				),
				Saver: c.Saver,
			},
		)
	default:
		return configuration, errors.New("unsupported resolver scheme")
	}
	// configure TLS
	if c.Config.TLSServerName != "" {
		configuration.HTTPConfig.TLSConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			ServerName: c.Config.TLSServerName,
		}
	}
	configuration.HTTPConfig.NoTLSVerify = c.Config.NoTLSVerify
	// configure proxy
	configuration.HTTPConfig.ProxyURL = c.ProxyURL
	return configuration, nil
}
