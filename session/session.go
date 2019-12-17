// Package session models a measurement session.
package session

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/m-lab/go/rtx"
	"github.com/ooni/probe-engine/bouncer"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/metadata"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/platform"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/resources"
)

// Session contains information on a measurement session.
type Session struct {
	// AssetsDir is the directory where to store assets.
	AssetsDir string

	// AvailableBouncers contains the available bouncers.
	AvailableBouncers []model.Service

	// AvailableCollectors contains the available collectors.
	AvailableCollectors []model.Service

	// AvailableTestHelpers contains the available test helpers.
	AvailableTestHelpers map[string][]model.Service

	// ExplicitProxy indicates that the user has explicitly
	// configured a proxy and wants us to know that. For more
	// info, see the documentation of New.
	ExplicitProxy bool

	// HTTPDefaultClient is the default HTTP client to use.
	HTTPDefaultClient *http.Client

	// HTTPNoProxyClient is a non-proxied HTTP client.
	HTTPNoProxyClient *http.Client

	// Logger is the log emitter.
	Logger log.Logger

	// PrivacySettings contains the collector privacy settings. The default
	// is to only redact the user's IP address from results.
	PrivacySettings model.PrivacySettings

	// SoftwareName contains the software name.
	SoftwareName string

	// SoftwareVersion contains the software version.
	SoftwareVersion string

	// TempDir is the directory where to store temporary files
	TempDir string

	// TLSConfig contains the TLS config
	TLSConfig *tls.Config

	// location is the probe location.
	location *model.LocationInfo
}

// New creates a new experiments session. The logger is the logger
// to use. The softwareName and softwareVersion identify the application
// that we're using. The assetsDir is the directory where assets will
// be downloaded and searched for. The proxy and tlsConfig arguments are
// more complicated as explained below.
//
// We configure two HTTP clients. The default client is used to
// contact services where the communication may be proxided. The
// non proxied client is for running measurements. The proxy
// argument only influences the former as the latter is not proxied
// in any case. (You can always edit the session after New has
// returned if you don't like this policy.)
//
// If proxy is nil, we'll configure the default client to use
// http.ProxyFromEnvironment. This means that we'll honour the
// HTTP_PROXY environment variable, if present. If proxy is
// non nil, we'll use it as a proxy unconditionally. Also, in
// the latter case, we'll keep track of the fact that we've
// an explicit proxy, and will do our best to avoid confusing
// services like mlab-ns, that may be confused by a proxy.
//
// Regarding tlsConfig, passing nil will always be a good idea,
// as in that case Go is more flexible in upgrading to http2, while
// with a custom CA it will not use http2. Yet, there may also be
// cases where a specific CA is still required. See the documention
// of httpx.NewTransport for more information on this.
func New(
	logger log.Logger, softwareName, softwareVersion, assetsDir string,
	proxy *url.URL, tlsConfig *tls.Config, tempDir string,
) *Session {
	return &Session{
		AssetsDir:     assetsDir,
		ExplicitProxy: proxy != nil,
		HTTPDefaultClient: httpx.NewTracingProxyingClient(
			logger, func(req *http.Request) (*url.URL, error) {
				if proxy != nil {
					return proxy, nil
				}
				return http.ProxyFromEnvironment(req)
			}, tlsConfig,
		),
		HTTPNoProxyClient: httpx.NewTracingProxyingClient(
			logger, nil, tlsConfig,
		),
		Logger: logger,
		PrivacySettings: model.PrivacySettings{
			IncludeCountry: true,
			IncludeASN:     true,
		},
		SoftwareName:    softwareName,
		SoftwareVersion: softwareVersion,
		TempDir:         tempDir,
		TLSConfig:       tlsConfig,
	}
}

// AddAvailableHTTPSBouncer adds the HTTP bouncer base URL to the list of URLs that are tried.
func (s *Session) AddAvailableHTTPSBouncer(baseURL string) {
	s.AvailableBouncers = append(s.AvailableBouncers, model.Service{
		Address: baseURL,
		Type:    "https",
	})
}

// AddAvailableHTTPSCollector adds an HTTP collector base URL to the list of URLs that are tried.
func (s *Session) AddAvailableHTTPSCollector(baseURL string) {
	s.AvailableCollectors = append(s.AvailableCollectors, model.Service{
		Address: baseURL,
		Type:    "https",
	})
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

// ResolverIP returns the resolver IP
func (s *Session) ResolverIP() string {
	ip := model.DefaultResolverIP
	if s.location != nil {
		ip = s.location.ResolverIP
	}
	return ip
}

func (s *Session) initOrchestraClient(
	ctx context.Context,
	clnt *orchestra.Client,
	maybeLogin func(ctx context.Context) error,
) (*orchestra.Client, error) {
	meta := metadata.Metadata{
		Platform:        platform.Name(),
		ProbeASN:        s.ProbeASNString(),
		ProbeCC:         s.ProbeCC(),
		SoftwareName:    s.SoftwareName,
		SoftwareVersion: s.SoftwareVersion,
		SupportedTests: []string{
			// TODO(bassosimone): do we need to declare more tests
			// here? I believe we're not using this functionality of
			// orchestra for now. Plus, we can always change later.
			"web_connectivity",
		},
	}
	if err := clnt.MaybeRegister(ctx, meta); err != nil {
		return nil, err
	}
	if err := maybeLogin(ctx); err != nil {
		return nil, err
	}
	return clnt, nil
}

// NewOrchestraClient creates a new orchestra client. This client is registered
// and logged in with the OONI orchestra. An error is returned on failure.
func (s *Session) NewOrchestraClient(ctx context.Context) (*orchestra.Client, error) {
	clnt := orchestra.NewClient(
		s.HTTPDefaultClient,
		s.Logger,
		s.UserAgent(),
		statefile.NewMemory(s.AssetsDir),
	)
	// Make sure we have location info so we can fill metadata
	if err := s.MaybeLookupLocation(ctx); err != nil {
		return nil, err
	}
	// TODO(bassosimone): until we implement persistent storage
	// it is advisable to use the testing service. The related
	// tracking GitHub issue is ooni/probe-engine#164.
	clnt.OrchestraBaseURL = "https://ps-test.ooni.io"
	clnt.RegistryBaseURL = "https://ps-test.ooni.io"
	return s.initOrchestraClient(
		ctx, clnt, clnt.MaybeLogin,
	)
}

func (s *Session) fetchResourcesIdempotent(ctx context.Context) error {
	return (&resources.Client{
		HTTPClient: s.HTTPDefaultClient, // proxy is OK
		Logger:     s.Logger,
		UserAgent:  s.UserAgent(),
		WorkDir:    s.AssetsDir,
	}).Ensure(ctx)
}

// ASNDatabasePath returns the path where the ASN database path should
// be if you have called s.FetchResourcesIdempotent.
func (s *Session) ASNDatabasePath() string {
	return filepath.Join(s.AssetsDir, resources.ASNDatabaseName)
}

// CABundlePath is like ASNDatabasePath but for the CA bundle path.
func (s *Session) CABundlePath() string {
	return filepath.Join(s.AssetsDir, resources.CABundleName)
}

// CountryDatabasePath is like ASNDatabasePath but for the country DB path.
func (s *Session) CountryDatabasePath() string {
	return filepath.Join(s.AssetsDir, resources.CountryDatabaseName)
}

func (s *Session) getAvailableBouncers() []model.Service {
	if len(s.AvailableBouncers) > 0 {
		return s.AvailableBouncers
	}
	return []model.Service{
		{
			Address: "https://bouncer.ooni.io",
			Type:    "https",
		},
	}
}

// UserAgent constructs the user agent to be used in this session.
func (s *Session) UserAgent() string {
	return s.SoftwareName + "/" + s.SoftwareVersion
}

func (s *Session) queryBouncer(
	ctx context.Context, query func(*bouncer.Client) error,
) error {
	for _, e := range s.getAvailableBouncers() {
		if e.Type != "https" {
			s.Logger.Debugf("session: unsupported bouncer type: %s", e.Type)
			continue
		}
		err := query(&bouncer.Client{
			BaseURL:    e.Address,
			HTTPClient: s.HTTPDefaultClient, // proxy is OK
			Logger:     s.Logger,
			UserAgent:  s.UserAgent(),
		})
		if err == nil {
			return nil
		}
		s.Logger.Warnf("session: bouncer error: %s", err.Error())
	}
	return errors.New("All available bouncers failed")
}

// MaybeLookupCollectors discovers collector information unless this bit of
// information has already been configured or discovered.
func (s *Session) MaybeLookupCollectors(ctx context.Context) error {
	if len(s.AvailableCollectors) > 0 {
		return nil
	}
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableCollectors, err = client.GetCollectors(ctx)
		return
	})
}

// MaybeLookupTestHelpers is like MaybeLookupCollectors for test helpers.
func (s *Session) MaybeLookupTestHelpers(ctx context.Context) error {
	if len(s.AvailableTestHelpers) > 0 {
		return nil
	}
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableTestHelpers, err = client.GetTestHelpers(ctx)
		return
	})
}

// MaybeLookupBackends discovers the available OONI backends. For each backend
// type, we query the bouncer only if we don't already have information.
//
// This is equivalent to calling MaybeLookupCollectors followed by calling
// MaybeLookupTestHelpers if MaybeLookupCollectors succeeds.
func (s *Session) MaybeLookupBackends(ctx context.Context) (err error) {
	err = s.MaybeLookupCollectors(ctx)
	if err != nil {
		return
	}
	err = s.MaybeLookupTestHelpers(ctx)
	return
}

func (s *Session) lookupProbeIP(ctx context.Context) (string, error) {
	return (&iplookup.Client{
		HTTPClient: s.HTTPNoProxyClient, // No proxy to have the correct IP
		Logger:     s.Logger,
		UserAgent:  s.UserAgent(),
	}).Do(ctx)
}

func (s *Session) lookupProbeNetwork(
	dbPath, probeIP string,
) (uint, string, error) {
	return mmdblookup.LookupASN(dbPath, probeIP, s.Logger)
}

func (s *Session) lookupProbeCC(
	dbPath, probeIP string,
) (string, error) {
	return mmdblookup.LookupCC(dbPath, probeIP, s.Logger)
}

func (s *Session) lookupResolverIP(ctx context.Context) (string, error) {
	return resolverlookup.First(ctx, nil)
}

// MaybeLookupLocation discovers details on the probe location only
// if this information it not already available.
func (s *Session) MaybeLookupLocation(ctx context.Context) (err error) {
	if s.location == nil {
		defer func() {
			if recover() != nil {
				// JUST KNOW WE'VE BEEN HERE
			}
		}()
		var (
			probeIP    string
			asn        uint
			org        string
			cc         string
			resolverIP string
		)
		err = s.fetchResourcesIdempotent(ctx)
		rtx.PanicOnError(err, "s.fetchResourcesIdempotent failed")
		probeIP, err = s.lookupProbeIP(ctx)
		rtx.PanicOnError(err, "s.lookupProbeIP failed")
		asn, org, err = s.lookupProbeNetwork(s.ASNDatabasePath(), probeIP)
		rtx.PanicOnError(err, "s.lookupProbeNetwork failed")
		cc, err = s.lookupProbeCC(s.CountryDatabasePath(), probeIP)
		rtx.PanicOnError(err, "s.lookupProbeCC failed")
		resolverIP, err = s.lookupResolverIP(ctx)
		rtx.PanicOnError(err, "s.lookupResolverIP failed")
		s.location = &model.LocationInfo{
			ASN:         asn,
			CountryCode: cc,
			NetworkName: org,
			ProbeIP:     probeIP,
			ResolverIP:  resolverIP,
		}
	}
	return
}
