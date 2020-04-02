package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/bouncer"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/internal/resources"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/modelx"
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
}

// Session is a measurement session
type Session struct {
	assetsDir            string
	availableBouncers    []model.Service
	availableCollectors  []model.Service
	availableTestHelpers map[string][]model.Service
	byteCounter          *dialer.ByteCounter
	httpDefaultClient    *http.Client
	httpNoProxyClient    *http.Client
	kvStore              model.KeyValueStore
	privacySettings      model.PrivacySettings
	explicitProxy        bool
	location             *model.LocationInfo
	logger               model.Logger
	queryBouncerCount    *atomicx.Int64
	softwareName         string
	softwareVersion      string
	tempDir              string
}

func newHTTPClient(sess *Session, proxy *url.URL, logger model.Logger) *http.Client {
	txp := netx.NewHTTPTransportWithProxyFunc(func(req *http.Request) (*url.URL, error) {
		if proxy != nil {
			return proxy, nil
		}
		return http.ProxyFromEnvironment(req)
	})
	return &http.Client{Transport: &sessHTTPTransport{
		beginning: time.Now(),
		logger:    logger,
		transport: txp,
	}}
}

type sessHTTPTransport struct {
	beginning time.Time
	logger    model.Logger
	transport *netx.HTTPTransport
}

func (t *sessHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.WithContext(modelx.WithMeasurementRoot(req.Context(), &modelx.MeasurementRoot{
		Beginning: t.beginning,
		Handler:   netxlogger.NewHandler(t.logger),
	}))
	return t.transport.RoundTrip(req)
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
		byteCounter: dialer.NewByteCounter(),
		kvStore:     config.KVStore,
		privacySettings: model.PrivacySettings{
			IncludeCountry: true,
			IncludeASN:     true,
		},
		explicitProxy:     config.ProxyURL != nil,
		logger:            config.Logger,
		queryBouncerCount: atomicx.NewInt64(),
		softwareName:      config.SoftwareName,
		softwareVersion:   config.SoftwareVersion,
		tempDir:           config.TempDir,
	}
	sess.httpDefaultClient = newHTTPClient(sess, config.ProxyURL, config.Logger)
	sess.httpNoProxyClient = newHTTPClient(sess, nil, config.Logger)
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

// KiBsReceived accounts for the KiBs received by the HTTP clients
// managed by this session so far, including experiments.
func (s *Session) KiBsReceived() float64 {
	return s.byteCounter.Received.Load()
}

// KiBsSent is like KiBsReceived but for the bytes sent.
func (s *Session) KiBsSent() float64 {
	return s.byteCounter.Sent.Load()
}

// CABundlePath is like ASNDatabasePath but for the CA bundle path.
func (s *Session) CABundlePath() string {
	return filepath.Join(s.assetsDir, resources.CABundleName)
}

// Close ensures that we close all the idle connections that the HTTP clients
// we are currently using may have created. Not calling this function may likely
// cause memory leaks in your application because of open idle connections.
func (s *Session) Close() error {
	s.httpDefaultClient.CloseIdleConnections()
	s.httpNoProxyClient.CloseIdleConnections()
	return nil
}

// CountryDatabasePath is like ASNDatabasePath but for the country DB path.
func (s *Session) CountryDatabasePath() string {
	return filepath.Join(s.assetsDir, resources.CountryDatabaseName)
}

// ExplicitProxy returns true if the user has explicitly set
// a proxy (as opposed to using the HTTP_PROXY envvar).
func (s *Session) ExplicitProxy() bool {
	return s.explicitProxy
}

// GetTestHelpersByName returns the available test helpers that
// use the specified name, or false if there's none.
func (s *Session) GetTestHelpersByName(name string) ([]model.Service, bool) {
	services, ok := s.availableTestHelpers[name]
	return services, ok
}

// DefaultHTTPClient returns the session's default HTTP client.
func (s *Session) DefaultHTTPClient() *http.Client {
	return s.httpDefaultClient
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

// NewExperimentBuilder returns a new experiment builder
// for the experiment with the given name, or an error if
// there's no such experiment with the given name
func (s *Session) NewExperimentBuilder(name string) (*ExperimentBuilder, error) {
	return newExperimentBuilder(s, name)
}

// NewOrchestraClient creates a new orchestra client. This client is registered
// and logged in with the OONI orchestra. An error is returned on failure.
func (s *Session) NewOrchestraClient(ctx context.Context) (model.ExperimentOrchestraClient, error) {
	ctx = dialer.WithSessionByteCounter(ctx, s.byteCounter)
	clnt := orchestra.NewClient(
		s.httpDefaultClient,
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

// UserAgent constructs the user agent to be used in this session.
func (s *Session) UserAgent() string {
	return s.softwareName + "/" + s.softwareVersion
}

func (s *Session) fetchResourcesIdempotent(ctx context.Context) error {
	ctx = dialer.WithSessionByteCounter(ctx, s.byteCounter)
	return (&resources.Client{
		HTTPClient: s.httpDefaultClient, // proxy is OK
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
	ctx = dialer.WithSessionByteCounter(ctx, s.byteCounter)
	return (&iplookup.Client{
		HTTPClient: s.httpNoProxyClient, // No proxy to have the correct IP
		Logger:     s.logger,
		UserAgent:  s.UserAgent(),
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
			resolverASN uint
			resolverIP  string
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
		resolverIP, err = s.lookupResolverIP(ctx)
		runtimex.PanicOnError(err, "s.lookupResolverIP failed")
		resolverASN, resolverOrg, err = s.lookupASN(
			s.ASNDatabasePath(), resolverIP,
		)
		runtimex.PanicOnError(err, "s.lookupASN #2 failed")
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
	ctx = dialer.WithSessionByteCounter(ctx, s.byteCounter)
	s.queryBouncerCount.Add(1)
	for _, e := range s.getAvailableBouncers() {
		if e.Type != "https" {
			s.logger.Debugf("session: unsupported bouncer type: %s", e.Type)
			continue
		}
		err := query(&bouncer.Client{
			BaseURL:    e.Address,
			HTTPClient: s.httpDefaultClient, // proxy is OK
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
