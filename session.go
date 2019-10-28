package engine

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"

	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/orchestra/testlists"
	"github.com/ooni/probe-engine/session"
)

// SessionConfig contains the Session config
type SessionConfig struct {
	AssetsDir       string
	Logger          log.Logger
	ProxyURL        *url.URL
	SoftwareName    string
	SoftwareVersion string
	TempDir         string
	TLSConfig       *tls.Config
}

// Session is a measurement session
type Session struct {
	session *session.Session
}

// NewSession creates a new session or returns an error
func NewSession(config SessionConfig) (*Session, error) {
	if config.AssetsDir == "" {
		return nil, errors.New("AssetsDir is empty")
	}
	if config.Logger == nil {
		return nil, errors.New("Logger is empty")
	}
	if config.SoftwareName == "" {
		return nil, errors.New("SoftwareName is empty")
	}
	if config.SoftwareVersion == "" {
		return nil, errors.New("SoftwareVersion is empty")
	}
	if config.TempDir == "" {
		return nil, errors.New("TempDir is empty")
	}
	sess := session.New(
		config.Logger,
		config.SoftwareName,
		config.SoftwareVersion,
		config.AssetsDir,
		config.ProxyURL,
		config.TLSConfig,
		config.TempDir,
	)
	return &Session{session: sess}, nil
}

// AddAvailableHTTPSBouncer adds an HTTPS bouncer to the list
// of bouncers that we'll try to contact.
func (sess *Session) AddAvailableHTTPSBouncer(baseURL string) {
	sess.session.AddAvailableHTTPSBouncer(baseURL)
}

// AddAvailableHTTPSCollector adds an HTTPS collector to the
// list of collectors that we'll try to use.
func (sess *Session) AddAvailableHTTPSCollector(baseURL string) {
	sess.session.AddAvailableHTTPSCollector(baseURL)
}

// MaybeLookupLocation is a caching location lookup call.
func (sess *Session) MaybeLookupLocation() error {
	return sess.session.MaybeLookupLocation(context.Background())
}

// MaybeLookupBackends is a caching OONI backends lookup call.
func (sess *Session) MaybeLookupBackends() error {
	return sess.session.MaybeLookupBackends(context.Background())
}

// ProbeASN returns the ASN of the probe's network.
func (sess *Session) ProbeASN() uint {
	return sess.session.ProbeASN()
}

// ProbeASNString is like ProbeASN but returns the "AS<%d>" string
// where the number is the number returned by ProbeASN.
func (sess *Session) ProbeASNString() string {
	return sess.session.ProbeASNString()
}

// ProbeCC returns the probe country code.
func (sess *Session) ProbeCC() string {
	return sess.session.ProbeCC()
}

// ProbeIP returns the probe IP.
func (sess *Session) ProbeIP() string {
	return sess.session.ProbeIP()
}

// ProbeNetworkName returns the probe network name.
func (sess *Session) ProbeNetworkName() string {
	return sess.session.ProbeNetworkName()
}

// ResolverIP returns the resolver IP.
func (sess *Session) ResolverIP() string {
	return sess.session.ResolverIP()
}

// SetIncludeProbeASN controls whether to include the ASN
func (sess *Session) SetIncludeProbeASN(value bool) {
	sess.session.PrivacySettings.IncludeASN = value
}

// SetIncludeProbeCC controls whether to include the country code
func (sess *Session) SetIncludeProbeCC(value bool) {
	sess.session.PrivacySettings.IncludeCountry = value
}

// SetIncludeProbeIP controls whether to include the IP
func (sess *Session) SetIncludeProbeIP(value bool) {
	sess.session.PrivacySettings.IncludeIP = value
}

// NewTestListsConfig returns prefilled settings for TestListsClient
// where in particular we have set the correct country code
func (sess *Session) NewTestListsConfig() *TestListsConfig {
	return &TestListsConfig{
		CountryCode: sess.session.ProbeCC(),
	}
}

// NewTestListsClient returns a new TestListsClient that is configured
// to perform requests in the context of this session
func (sess *Session) NewTestListsClient() *TestListsClient {
	return &TestListsClient{
		client: testlists.NewClient(sess.session),
	}
}

// NewExperimentBuilder returns a new experiment builder
// for the experiment with the given name, or an error if
// there's no such experiment with the given name
func (sess *Session) NewExperimentBuilder(name string) (*ExperimentBuilder, error) {
	return newExperimentBuilder(sess, name)
}
