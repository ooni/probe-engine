package urlgetter

import (
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/httptransport"
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
	HTTPConfig httptransport.Config
	DNSClient  httptransport.DNSClient
}

// CloseIdleConnections will close idle connections, if needed.
func (c Configuration) CloseIdleConnections() {
	c.DNSClient.CloseIdleConnections()
}

// NewConfiguration builds a new measurement configuration.
func (c Configurer) NewConfiguration() (Configuration, error) {
	// set up defaults
	configuration := Configuration{
		HTTPConfig: httptransport.Config{
			BogonIsError:        c.Config.RejectDNSBogons,
			CacheResolutions:    true,
			ContextByteCounting: true,
			DialSaver:           c.Saver,
			HTTPSaver:           c.Saver,
			Logger:              c.Logger,
			ReadWriteSaver:      c.Saver,
			ResolveSaver:        c.Saver,
			TLSSaver:            c.Saver,
		},
	}
	// fill DNS cache
	if c.Config.DNSCache != "" {
		entry := strings.Split(c.Config.DNSCache, " ")
		if len(entry) < 2 {
			return configuration, errors.New("invalid DNSCache string")
		}
		domainregex := regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
		if !domainregex.MatchString(entry[0]) {
			return configuration, errors.New("invalid domain in DNSCache")
		}
		var addresses []string
		for i := 1; i < len(entry); i++ {
			if net.ParseIP(entry[i]) == nil {
				return configuration, errors.New("invalid IP in DNSCache")
			}
			addresses = append(addresses, entry[i])
		}
		configuration.HTTPConfig.DNSCache = map[string][]string{
			entry[0]: addresses,
		}
	}
	dnsclient, err := httptransport.NewDNSClient(
		configuration.HTTPConfig, c.Config.ResolverURL,
	)
	if err != nil {
		return configuration, err
	}
	configuration.DNSClient = dnsclient
	configuration.HTTPConfig.BaseResolver = dnsclient.Resolver
	// configure TLS
	configuration.HTTPConfig.TLSConfig = &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
	}
	if c.Config.TLSServerName != "" {
		configuration.HTTPConfig.TLSConfig.ServerName = c.Config.TLSServerName
	}
	switch c.Config.TLSVersion {
	case "TLSv1.3":
		configuration.HTTPConfig.TLSConfig.MinVersion = tls.VersionTLS13
		configuration.HTTPConfig.TLSConfig.MaxVersion = tls.VersionTLS13
	case "TLSv1.2":
		configuration.HTTPConfig.TLSConfig.MinVersion = tls.VersionTLS12
		configuration.HTTPConfig.TLSConfig.MaxVersion = tls.VersionTLS12
	case "TLSv1.1":
		configuration.HTTPConfig.TLSConfig.MinVersion = tls.VersionTLS11
		configuration.HTTPConfig.TLSConfig.MaxVersion = tls.VersionTLS11
	case "TLSv1.0", "TLSv1":
		configuration.HTTPConfig.TLSConfig.MinVersion = tls.VersionTLS10
		configuration.HTTPConfig.TLSConfig.MaxVersion = tls.VersionTLS10
	case "":
		// nothing
	default:
		return configuration, errors.New("unsupported TLS version")
	}
	configuration.HTTPConfig.NoTLSVerify = c.Config.NoTLSVerify
	// configure proxy
	configuration.HTTPConfig.ProxyURL = c.ProxyURL
	return configuration, nil
}
