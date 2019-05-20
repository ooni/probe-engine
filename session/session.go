// Package session models a measurement session.
package session

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ooni/probe-engine/bouncer"
	"github.com/ooni/probe-engine/geoiplookup/constants"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/resources"
)

// Session contains information on a measurement session.
type Session struct {
	// AvailableBouncers contains the available bouncers.
	AvailableBouncers []model.Service

	// AvailableCollectors contains the available collectors.
	AvailableCollectors []model.Service

	// AvailableTestHelpers contains the available test helpers.
	AvailableTestHelpers map[string][]model.Service

	// HTTPDefaultClient is the default HTTP client to use.
	HTTPDefaultClient *http.Client

	// HTTPNoProxyClient is a non-proxied HTTP client.
	HTTPNoProxyClient *http.Client

	// Logger is the log emitter.
	Logger log.Logger

	// ProbeASN contains the probe ASN.
	ProbeASN string

	// ProbeCC contains the probe CC.
	ProbeCC string

	// ProbeIP contains the probe IP.
	ProbeIP string

	// ProbeNetworkName contains the probe network name.
	ProbeNetworkName string

	// ResolverIP is the resolver's IP.
	ResolverIP string

	// SoftwareName contains the software name.
	SoftwareName string

	// SoftwareVersion contains the software version.
	SoftwareVersion string

	// WorkDir is the session's working directory.
	WorkDir string
}

// New creates a new measurements session.
func New(logger log.Logger, softwareName, softwareVersion string) *Session {
	return &Session{
		HTTPDefaultClient: httpx.NewTracingProxyingClient(
			logger, http.ProxyFromEnvironment,
		),
		HTTPNoProxyClient: httpx.NewTracingProxyingClient(logger, nil),
		Logger:            logger,
		ProbeASN:          constants.DefaultProbeASN,
		ProbeCC:           constants.DefaultProbeCC,
		ProbeIP:           constants.DefaultProbeIP,
		ProbeNetworkName:  constants.DefaultProbeNetworkName,
		ResolverIP:        constants.DefaultResolverIP,
		SoftwareName:      softwareName,
		SoftwareVersion:   softwareVersion,
		WorkDir:           os.TempDir(),
	}
}

// FetchResourcesIdempotent makes sure we have the resources we
// need for running inside of s.WorkDir.
func (s *Session) FetchResourcesIdempotent(ctx context.Context) error {
	return (&resources.Client{
		HTTPClient: s.HTTPDefaultClient, // proxy is OK
		Logger:     s.Logger,
		UserAgent:  s.UserAgent(),
		WorkDir:    s.WorkDir,
	}).Ensure(ctx)
}

// ASNDatabasePath returns the path where the ASN database path should
// be if you have called s.FetchResourcesIdempotent.
func (s *Session) ASNDatabasePath() string {
	return filepath.Join(s.WorkDir, resources.ASNDatabaseName)
}

// CABundlePath is like ASNDatabasePath but for the CA bundle path.
func (s *Session) CABundlePAth() string {
	return filepath.Join(s.WorkDir, resources.CABundleName)
}

// CountryDatabasePath is like ASNDatabasePath but for the country DB path.
func (s *Session) CountryDatabasePath() string {
	return filepath.Join(s.WorkDir, resources.CountryDatabaseName)
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

// LookupCollectors discovers the available collectors.
func (s *Session) LookupCollectors(ctx context.Context) error {
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableCollectors, err = client.GetCollectors(ctx)
		return
	})
}

// LookupTestHelpers discovers the available test helpers.
func (s *Session) LookupTestHelpers(ctx context.Context) error {
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableTestHelpers, err = client.GetTestHelpers(ctx)
		return
	})
}

// LookupProbeIP discovers the probe IP.
func (s *Session) LookupProbeIP(ctx context.Context) (err error) {
	s.ProbeIP, err = (&iplookup.Client{
		HTTPClient: s.HTTPNoProxyClient, // No proxy to have the right IP
		Logger:     s.Logger,
		UserAgent:  s.UserAgent(),
	}).Do(ctx)
	s.Logger.Debugf("ProbeIP: %s", s.ProbeIP)
	return
}

// LookupProbeASN discovers the probe ASN.
func (s *Session) LookupProbeASN(databasePath string) (err error) {
	if s.ProbeASN == constants.DefaultProbeASN {
		s.ProbeASN, s.ProbeNetworkName, err = mmdblookup.LookupASN(
			databasePath, s.ProbeIP,
		)
	}
	s.Logger.Debugf("ProbeASN: %s", s.ProbeASN)
	return
}

// LookupProbeCC discovers the probe CC.
func (s *Session) LookupProbeCC(databasePath string) (err error) {
	if s.ProbeCC == constants.DefaultProbeCC {
		s.ProbeCC, err = mmdblookup.LookupCC(databasePath, s.ProbeIP)
	}
	s.Logger.Debugf("ProbeCC: %s", s.ProbeCC)
	return
}

// LookupProbeNetworkName discovers the probe network name.
func (s *Session) LookupProbeNetworkName(databasePath string) (err error) {
	if s.ProbeNetworkName == constants.DefaultProbeNetworkName {
		s.ProbeASN, s.ProbeNetworkName, err = mmdblookup.LookupASN(
			databasePath, s.ProbeIP,
		)
	}
	s.Logger.Debugf("ProbeNetworkName: %s", s.ProbeNetworkName)
	return
}

// LookupResolverIP discovers the resolver IP.
func (s *Session) LookupResolverIP(ctx context.Context) error {
	if s.ResolverIP != constants.DefaultResolverIP {
		return nil
	}
	addrs, err := resolverlookup.Do(ctx, nil)
	if err != nil {
		return err
	}
	if len(addrs) < 1 {
		return errors.New("No resolver IPs returned")
	}
	s.ResolverIP = addrs[0]
	s.Logger.Debugf("ResolverIP: %s", s.ResolverIP)
	return nil
}
