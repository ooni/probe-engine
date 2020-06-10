package engine

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/internal/resources"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/internal/sessionresolver"
	"github.com/ooni/probe-engine/internal/sessiontunnel"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/version"
)

// SessionConfig contains the Session config
type SessionConfig struct {
	AssetsDir              string
	AvailableProbeServices []model.Service
	KVStore                KVStore
	Logger                 model.Logger
	PrivacySettings        model.PrivacySettings
	ProxyURL               *url.URL
	SoftwareName           string
	SoftwareVersion        string
	TempDir                string
	TorArgs                []string
	TorBinary              string
}

// Session is a measurement session
type Session struct {
	assetsDir               string
	availableProbeServices  []model.Service
	availableTestHelpers    map[string][]model.Service
	byteCounter             *bytecounter.Counter
	httpDefaultTransport    httptransport.RoundTripper
	kvStore                 model.KeyValueStore
	privacySettings         model.PrivacySettings
	location                *model.LocationInfo
	logger                  model.Logger
	proxyURL                *url.URL
	queryProbeServicesCount *atomicx.Int64
	resolver                *sessionresolver.Resolver
	selectedProbeService    *model.Service
	softwareName            string
	softwareVersion         string
	tempDir                 string
	torArgs                 []string
	torBinary               string
	tunnelMu                sync.Mutex
	tunnelName              string
	tunnel                  sessiontunnel.Tunnel
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
	if config.KVStore == nil {
		config.KVStore = kvstore.NewMemoryKeyValueStore()
	}
	// Implementation note: if config.TempDir is empty, then Go will
	// use the temporary directory on the current system. This should
	// work on Desktop. We tested that it did also work on iOS, but
	// we have also seen on 2020-06-10 that it does not work on Android.
	tempDir, err := ioutil.TempDir(config.TempDir, "ooniengine")
	if err != nil {
		return nil, err
	}
	sess := &Session{
		assetsDir:               config.AssetsDir,
		availableProbeServices:  config.AvailableProbeServices,
		byteCounter:             bytecounter.New(),
		kvStore:                 config.KVStore,
		privacySettings:         config.PrivacySettings,
		logger:                  config.Logger,
		proxyURL:                config.ProxyURL,
		queryProbeServicesCount: atomicx.NewInt64(),
		softwareName:            config.SoftwareName,
		softwareVersion:         config.SoftwareVersion,
		tempDir:                 tempDir,
		torArgs:                 config.TorArgs,
		torBinary:               config.TorBinary,
	}
	httpConfig := httptransport.Config{
		ByteCounter:  sess.byteCounter,
		BogonIsError: true,
		Logger:       sess.logger,
	}
	sess.resolver = sessionresolver.New(httpConfig)
	httpConfig.FullResolver = sess.resolver
	httpConfig.ProxyURL = config.ProxyURL // no need to proxy the resolver
	sess.httpDefaultTransport = httptransport.New(httpConfig)
	return sess, nil
}

// ASNDatabasePath returns the path where the ASN database path should
// be if you have called s.FetchResourcesIdempotent.
func (s *Session) ASNDatabasePath() string {
	return filepath.Join(s.assetsDir, resources.ASNDatabaseName)
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
// we are currently using may have created. It will also remove the temp dir
// that contains data from this session. Not calling this function may likely
// cause memory leaks in your application because of open idle connections,
// as well as excessive usage of disk space.
func (s *Session) Close() error {
	s.httpDefaultTransport.CloseIdleConnections()
	s.resolver.CloseIdleConnections()
	if s.tunnel != nil {
		s.tunnel.Stop()
	}
	return os.RemoveAll(s.tempDir)
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

// KeyValueStore returns the configured key-value store.
func (s *Session) KeyValueStore() model.KeyValueStore {
	return s.kvStore
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

// ErrAlreadyUsingProxy indicates that we cannot create a tunnel with
// a specific name because we already configured a proxy.
var ErrAlreadyUsingProxy = errors.New(
	"session: cannot create a new tunnel of this kind: we are already using a proxy",
)

// MaybeStartTunnel starts the requested tunnel.
//
// This function silently succeeds if we're already using a tunnel with
// the same name or if the requested tunnel name is the empty string. This
// function fails, tho, when we already have a proxy or a tunnel with
// another name and we try to open a tunnel. This function of course also
// fails if we cannot start the requested tunnel. All in all, if you request
// for a tunnel name that is not the empty string and you get a nil error,
// you can be confident that session.ProxyURL() gives you the tunnel URL.
//
// The tunnel will be closed by session.Close().
func (s *Session) MaybeStartTunnel(ctx context.Context, name string) error {
	s.tunnelMu.Lock()
	defer s.tunnelMu.Unlock()
	if s.tunnel != nil && s.tunnelName == name {
		// We've been asked more than once to start the same tunnel.
		return nil
	}
	if s.proxyURL != nil && name == "" {
		// The user configured a proxy and here we're not actually trying
		// to start any tunnel since `name` is empty.
		return nil
	}
	if s.proxyURL != nil || s.tunnel != nil {
		// We already have a proxy or we have a different tunnel. Because a tunnel
		// sets a proxy, the second check for s.tunnel is for robustness.
		return ErrAlreadyUsingProxy
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
	if tunnel == nil {
		return nil
	}
	s.tunnelName = name
	s.tunnel = tunnel
	s.proxyURL = tunnel.SOCKS5ProxyURL()
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
		orchestra.NewStateFile(s.kvStore),
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
func (s *Session) UserAgent() (useragent string) {
	useragent += s.softwareName + "/" + s.softwareVersion
	useragent += " ooniprobe-engine/" + version.Version
	return
}

func (s *Session) fetchResourcesIdempotent(ctx context.Context) error {
	return (&resources.Client{
		HTTPClient: s.DefaultHTTPClient(),
		Logger:     s.logger,
		UserAgent:  s.UserAgent(),
		WorkDir:    s.assetsDir,
	}).Ensure(ctx)
}

func (s *Session) getAvailableProbeServices() []model.Service {
	if len(s.availableProbeServices) > 0 {
		return s.availableProbeServices
	}
	return probeservices.Default()
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
	return mmdblookup.ASN(dbPath, ip)
}

func (s *Session) lookupProbeIP(ctx context.Context) (string, error) {
	return (&iplookup.Client{
		HTTPClient: s.DefaultHTTPClient(),
		Logger:     s.logger,
		UserAgent:  httpheader.RandomUserAgent(), // no need to identify as OONI
	}).Do(ctx)
}

func (s *Session) lookupProbeCC(dbPath, probeIP string) (string, error) {
	return mmdblookup.CC(dbPath, probeIP)
}

func (s *Session) lookupResolverIP(ctx context.Context) (string, error) {
	return resolverlookup.First(ctx, nil)
}

func (s *Session) maybeLookupBackends(ctx context.Context) error {
	// TODO(bassosimone): do we need a mutex here?
	if s.selectedProbeService != nil {
		return nil
	}
	s.queryProbeServicesCount.Add(1)
	candidates := probeservices.TryAll(ctx, s, s.getAvailableProbeServices())
	selected := probeservices.SelectBest(candidates)
	if selected == nil {
		return errors.New("all available probe services failed")
	}
	s.logger.Infof("session: using probe services: %+v", selected.Endpoint)
	s.selectedProbeService = &selected.Endpoint
	s.availableTestHelpers = selected.TestHelpers
	return nil
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

var _ model.ExperimentSession = &Session{}
