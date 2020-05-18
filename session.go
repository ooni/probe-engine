package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/bouncer"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/internal/resources"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/internal/sessiontunnel"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
)

// SessionConfig contains the Session config
type SessionConfig struct {
	AssetsDir       string
	KVStore         KVStore
	Logger          model.Logger
	ProxyURL        *url.URL
	SoftwareName    string
	SoftwareVersion string
	TempDir         string
	TorArgs         []string
	TorBinary       string
}

// Session is a measurement session
type Session struct {
	assetsDir            string
	availableBouncers    []model.Service
	availableCollectors  []model.Service
	availableTestHelpers map[string][]model.Service
	byteCounter          *bytecounter.Counter
	httpDefaultTransport httptransport.RoundTripper
	kvStore              model.KeyValueStore
	privacySettings      model.PrivacySettings
	location             *model.LocationInfo
	logger               model.Logger
	proxyURL             *url.URL
	queryBouncerCount    *atomicx.Int64
	softwareName         string
	softwareVersion      string
	tempDir              string
	torArgs              []string
	torBinary            string
	tunnel               sessiontunnel.Tunnel
	tunnelMu             sync.Mutex
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
	sess := &Session{
		assetsDir:   config.AssetsDir,
		byteCounter: bytecounter.New(),
		kvStore:     config.KVStore,
		privacySettings: model.PrivacySettings{
			IncludeCountry: true,
			IncludeASN:     true,
		},
		logger:            config.Logger,
		proxyURL:          config.ProxyURL,
		queryBouncerCount: atomicx.NewInt64(),
		softwareName:      config.SoftwareName,
		softwareVersion:   config.SoftwareVersion,
		tempDir:           config.TempDir,
		torArgs:           config.TorArgs,
		torBinary:         config.TorBinary,
	}
	sess.httpDefaultTransport = httptransport.New(httptransport.Config{
		ByteCounter:  sess.byteCounter,
		BogonIsError: true,
		Logger:       sess.logger,
		ProxyURL:     config.ProxyURL,
	})
	return sess, nil
}

// ASNDatabasePath returns the path where the ASN database path should
// be if you have called s.FetchResourcesIdempotent.
func (s *Session) ASNDatabasePath() string {
	return filepath.Join(s.assetsDir, resources.ASNDatabaseName)
}

// AddAvailableHTTPSBouncer adds an HTTPS bouncer to the list
// of bouncers that we'll try to contact.
func (s *Session) AddAvailableHTTPSBouncer(baseURL string) {
	s.availableBouncers = append(s.availableBouncers, model.Service{
		Address: baseURL,
		Type:    "https",
	})
}

// AddAvailableHTTPSCollector adds an HTTPS collector to the
// list of collectors that we'll try to use.
func (s *Session) AddAvailableHTTPSCollector(baseURL string) {
	s.availableCollectors = append(s.availableCollectors, model.Service{
		Address: baseURL,
		Type:    "https",
	})
}

// KibiBytesReceived accounts for the KibiBytes received by the HTTP clients
// managed by this session so far, including experiments.
func (s *Session) KibiBytesReceived() float64 {
	return s.byteCounter.KibiBytesReceived()
}

// KibiBytesSent is like KibiBytesReceived but for the bytes sent.
func (s *Session) KibiBytesSent() float64 {
	return s.byteCounter.KibiBytesSent()
}

// CABundlePath is like ASNDatabasePath but for the CA bundle path.
func (s *Session) CABundlePath() string {
	return filepath.Join(s.assetsDir, resources.CABundleName)
}

// Close ensures that we close all the idle connections that the HTTP clients
// we are currently using may have created. Not calling this function may likely
// cause memory leaks in your application because of open idle connections.
func (s *Session) Close() error {
	s.httpDefaultTransport.CloseIdleConnections()
	if s.tunnel != nil {
		s.tunnel.Stop()
	}
	return nil
}

// CountryDatabasePath is like ASNDatabasePath but for the country DB path.
func (s *Session) CountryDatabasePath() string {
	return filepath.Join(s.assetsDir, resources.CountryDatabaseName)
}

// GetTestHelpersByName returns the available test helpers that
// use the specified name, or false if there's none.
func (s *Session) GetTestHelpersByName(name string) ([]model.Service, bool) {
	services, ok := s.availableTestHelpers[name]
	return services, ok
}

// DefaultHTTPClient returns the session's default HTTP client.
func (s *Session) DefaultHTTPClient() *http.Client {
	return &http.Client{Transport: s.httpDefaultTransport}
}

// Logger returns the logger used by the session.
func (s *Session) Logger() model.Logger {
	return s.logger
}

// MaybeLookupLocation is a caching location lookup call.
func (s *Session) MaybeLookupLocation() error {
	return s.maybeLookupLocation(context.Background())
}

// MaybeLookupBackends is a caching OONI backends lookup call.
func (s *Session) MaybeLookupBackends() error {
	return s.maybeLookupBackends(context.Background())
}

// MaybeStartTunnel starts the requested tunnel. This function silently
// succeeds if we're already using a tunnel (or proxy) or if the provided
// tunnel name is the empty string (i.e. no tunnel). This function fails
// if we don't know the requested tunnel, or if starting the tunnel actually
// fails. We currently only know the "psiphon" tunnel. A side effect of
// starting the tunnel is that we will correctly set the proxy URL. Note
// that the tunnel will be active until session.Close is called.
func (s *Session) MaybeStartTunnel(ctx context.Context, name string) error {
	s.tunnelMu.Lock()
	defer s.tunnelMu.Unlock()
	if s.proxyURL != nil || s.tunnel != nil {
		s.logger.Debugf("not starting tunnel because we already have a proxy")
		return nil
	}
	tunnel, err := sessiontunnel.Start(ctx, sessiontunnel.Config{
		Name:    name,
		Session: s,
	})
	if err != nil {
		s.logger.Warnf("cannot start tunnel: %+v", err)
		return err
	}
	// Implementation note: tunnel _may_ be NIL here if name is ""
	if tunnel != nil {
		s.tunnel = tunnel
		s.proxyURL = tunnel.SOCKS5ProxyURL()
	}
	return nil
}

// NewExperimentBuilder returns a new experiment builder
// for the experiment with the given name, or an error if
// there's no such experiment with the given name
func (s *Session) NewExperimentBuilder(name string) (*ExperimentBuilder, error) {
	return newExperimentBuilder(s, name)
}

// NewOrchestraClient creates a new orchestra client. This client is registered
// and logged in with the OONI orchestra. An error is returned on failure.
func (s *Session) NewOrchestraClient(ctx context.Context) (model.ExperimentOrchestraClient, error) {
	clnt := orchestra.NewClient(
		s.DefaultHTTPClient(),
		s.logger,
		s.UserAgent(),
		statefile.New(s.kvStore),
	)
	return s.initOrchestraClient(
		ctx, clnt, clnt.MaybeLogin,
	)
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
func (s *Session) Platform() string {
	return platform.Name()
}

// ProbeASNString returns the probe ASN as a string.
func (s *Session) ProbeASNString() string {
	return fmt.Sprintf("AS%d", s.ProbeASN())
}

// ProbeASN returns the probe ASN as an integer.
func (s *Session) ProbeASN() uint {
	asn := model.DefaultProbeASN
	if s.location != nil {
		asn = s.location.ASN
	}
	return asn
}

// ProbeCC returns the probe CC.
func (s *Session) ProbeCC() string {
	cc := model.DefaultProbeCC
	if s.location != nil {
		cc = s.location.CountryCode
	}
	return cc
}

// ProbeNetworkName returns the probe network name.
func (s *Session) ProbeNetworkName() string {
	nn := model.DefaultProbeNetworkName
	if s.location != nil {
		nn = s.location.NetworkName
	}
	return nn
}

// ProbeIP returns the probe IP.
func (s *Session) ProbeIP() string {
	ip := model.DefaultProbeIP
	if s.location != nil {
		ip = s.location.ProbeIP
	}
	return ip
}

// ProxyURL returns the Proxy URL, or nil if not set
func (s *Session) ProxyURL() *url.URL {
	return s.proxyURL
}

// ResolverASNString returns the resolver ASN as a string
func (s *Session) ResolverASNString() string {
	return fmt.Sprintf("AS%d", s.ResolverASN())
}

// ResolverASN returns the resolver ASN
func (s *Session) ResolverASN() uint {
	asn := model.DefaultResolverASN
	if s.location != nil {
		asn = s.location.ResolverASN
	}
	return asn
}

// ResolverIP returns the resolver IP
func (s *Session) ResolverIP() string {
	ip := model.DefaultResolverIP
	if s.location != nil {
		ip = s.location.ResolverIP
	}
	return ip
}

// ResolverNetworkName returns the resolver network name.
func (s *Session) ResolverNetworkName() string {
	nn := model.DefaultResolverNetworkName
	if s.location != nil {
		nn = s.location.ResolverNetworkName
	}
	return nn
}

// SetIncludeProbeASN controls whether to include the ASN
func (s *Session) SetIncludeProbeASN(value bool) {
	s.privacySettings.IncludeASN = value
}

// SetIncludeProbeCC controls whether to include the country code
func (s *Session) SetIncludeProbeCC(value bool) {
	s.privacySettings.IncludeCountry = value
}

// SetIncludeProbeIP controls whether to include the IP
func (s *Session) SetIncludeProbeIP(value bool) {
	s.privacySettings.IncludeIP = value
}

// SoftwareName returns the application name.
func (s *Session) SoftwareName() string {
	return s.softwareName
}

// SoftwareVersion returns the application version.
func (s *Session) SoftwareVersion() string {
	return s.softwareVersion
}

// TempDir returns the temporary directory.
func (s *Session) TempDir() string {
	return s.tempDir
}

// TorArgs returns the configured extra args for the tor binary. If not set
// we will not pass in any extra arg. Applies to `-OTunnel=tor` mainly.
func (s *Session) TorArgs() []string {
	return s.torArgs
}

// TorBinary returns the configured path to the tor binary. If not set
// we will attempt to use "tor". Applies to `-OTunnel=tor` mainly.
func (s *Session) TorBinary() string {
	return s.torBinary
}

// TunnelBootstrapTime returns the time required to bootstrap the tunnel
// we're using, or zero if we're using no tunnel.
func (s *Session) TunnelBootstrapTime() time.Duration {
	if s.tunnel == nil {
		return 0
	}
	return s.tunnel.BootstrapTime()
}

// UserAgent constructs the user agent to be used in this session.
func (s *Session) UserAgent() string {
	return s.softwareName + "/" + s.softwareVersion
}

func (s *Session) fetchResourcesIdempotent(ctx context.Context) error {
	return (&resources.Client{
		HTTPClient: s.DefaultHTTPClient(),
		Logger:     s.logger,
		UserAgent:  s.UserAgent(),
		WorkDir:    s.assetsDir,
	}).Ensure(ctx)
}

func (s *Session) getAvailableBouncers() []model.Service {
	if len(s.availableBouncers) > 0 {
		return s.availableBouncers
	}
	return []model.Service{{
		Address: "https://bouncer.ooni.io",
		Type:    "https",
	}}
}

func (s *Session) initOrchestraClient(
	ctx context.Context, clnt *orchestra.Client,
	maybeLogin func(ctx context.Context) error,
) (*orchestra.Client, error) {
	// The original implementation has as its only use case that we
	// were registering and logging in for sending an update regarding
	// the probe whereabouts. Yet here in probe-engine, the orchestra
	// is currently only used to fetch inputs. For this purpose, we don't
	// need to communicate any specific information. The code that will
	// perform an update should be responsible of doing that.
	meta := metadata.Metadata{
		Platform:        "miniooni",
		ProbeASN:        "AS0",
		ProbeCC:         "ZZ",
		SoftwareName:    "miniooni",
		SoftwareVersion: "0.1.0-dev",
		SupportedTests:  []string{"web_connectivity"},
	}
	if err := clnt.MaybeRegister(ctx, meta); err != nil {
		return nil, err
	}
	if err := maybeLogin(ctx); err != nil {
		return nil, err
	}
	return clnt, nil
}

func (s *Session) lookupASN(dbPath, ip string) (uint, string, error) {
	return mmdblookup.LookupASN(dbPath, ip, s.logger)
}

func (s *Session) lookupProbeIP(ctx context.Context) (string, error) {
	return (&iplookup.Client{
		HTTPClient: s.DefaultHTTPClient(),
		Logger:     s.logger,
		UserAgent:  httpheader.RandomUserAgent(), // no need to identify as OONI
	}).Do(ctx)
}

func (s *Session) lookupProbeCC(dbPath, probeIP string) (string, error) {
	return mmdblookup.LookupCC(dbPath, probeIP, s.logger)
}

func (s *Session) lookupResolverIP(ctx context.Context) (string, error) {
	return resolverlookup.First(ctx, nil)
}

func (s *Session) maybeLookupBackends(ctx context.Context) (err error) {
	err = s.maybeLookupCollectors(ctx)
	if err != nil {
		return
	}
	err = s.maybeLookupTestHelpers(ctx)
	return
}

func (s *Session) maybeLookupCollectors(ctx context.Context) error {
	if len(s.availableCollectors) > 0 {
		return nil
	}
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.availableCollectors, err = client.GetCollectors(ctx)
		return
	})
}

func (s *Session) maybeLookupLocation(ctx context.Context) (err error) {
	if s.location == nil {
		defer func() {
			if recover() != nil {
				// JUST KNOW WE'VE BEEN HERE
			}
		}()
		var (
			probeIP     string
			asn         uint
			org         string
			cc          string
			resolverASN uint   = model.DefaultResolverASN
			resolverIP  string = model.DefaultResolverIP
			resolverOrg string
		)
		err = s.fetchResourcesIdempotent(ctx)
		runtimex.PanicOnError(err, "s.fetchResourcesIdempotent failed")
		probeIP, err = s.lookupProbeIP(ctx)
		runtimex.PanicOnError(err, "s.lookupProbeIP failed")
		asn, org, err = s.lookupASN(s.ASNDatabasePath(), probeIP)
		runtimex.PanicOnError(err, "s.lookupASN #1 failed")
		cc, err = s.lookupProbeCC(s.CountryDatabasePath(), probeIP)
		runtimex.PanicOnError(err, "s.lookupProbeCC failed")
		if s.proxyURL == nil {
			resolverIP, err = s.lookupResolverIP(ctx)
			runtimex.PanicOnError(err, "s.lookupResolverIP failed")
			resolverASN, resolverOrg, err = s.lookupASN(
				s.ASNDatabasePath(), resolverIP,
			)
			runtimex.PanicOnError(err, "s.lookupASN #2 failed")
		}
		s.location = &model.LocationInfo{
			ASN:                 asn,
			CountryCode:         cc,
			NetworkName:         org,
			ProbeIP:             probeIP,
			ResolverASN:         resolverASN,
			ResolverIP:          resolverIP,
			ResolverNetworkName: resolverOrg,
		}
	}
	return
}

func (s *Session) maybeLookupTestHelpers(ctx context.Context) error {
	if len(s.availableTestHelpers) > 0 {
		return nil
	}
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.availableTestHelpers, err = client.GetTestHelpers(ctx)
		return
	})
}

func (s *Session) queryBouncer(ctx context.Context, query func(*bouncer.Client) error) error {
	s.queryBouncerCount.Add(1)
	for _, e := range s.getAvailableBouncers() {
		if e.Type != "https" {
			s.logger.Debugf("session: unsupported bouncer type: %s", e.Type)
			continue
		}
		err := query(&bouncer.Client{
			BaseURL:    e.Address,
			HTTPClient: s.DefaultHTTPClient(),
			Logger:     s.logger,
			UserAgent:  s.UserAgent(),
		})
		if err == nil {
			return nil
		}
		s.logger.Warnf("session: bouncer error: %s", err.Error())
	}
	return errors.New("All available bouncers failed")
}

var _ model.ExperimentSession = &Session{}
