package engine

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"

	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/session"
)

// SessionConfig contains the Session config
type SessionConfig struct {
	AssetsDir       string
	Logger          log.Logger
	KVStore         KVStore
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
	if config.KVStore == nil {
		config.KVStore = kvstore.NewMemoryKeyValueStore()
	}
	sess := session.New(
		config.Logger,
		config.SoftwareName,
		config.SoftwareVersion,
		config.AssetsDir,
		config.ProxyURL,
		config.TLSConfig,
		config.TempDir,
		config.KVStore,
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

// Platform returns the current platform. The platform is one of:
//
// - android
// - ios
// - linux
// - macos
// - windows
// - unknown
//
// When running on the iOS simulator, the returned platform is
// macos rather than ios if CGO is disabled. This is a known issue,
// that however should have a very limited impact.
func (sess *Session) Platform() string {
	return platform.Name()
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

// ResolverASNString returns the resolver ASN as as string
func (sess *Session) ResolverASNString() string {
	return sess.session.ResolverASNString()
}

// ResolverASN returns the resolver ASN
func (sess *Session) ResolverASN() uint {
	return sess.session.ResolverASN()
}

// ResolverIP returns the resolver IP.
func (sess *Session) ResolverIP() string {
	return sess.session.ResolverIP()
}

// ResolverNetworkName returns the resolver network name
func (sess *Session) ResolverNetworkName() string {
	return sess.session.ResolverNetworkName()
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

// NewExperimentBuilder returns a new experiment builder
// for the experiment with the given name, or an error if
// there's no such experiment with the given name
func (sess *Session) NewExperimentBuilder(name string) (*ExperimentBuilder, error) {
	return newExperimentBuilder(sess, name)
}
