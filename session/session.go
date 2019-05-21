// Package session models a measurement session.
package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ooni/probe-engine/bouncer"
	"github.com/ooni/probe-engine/geoiplookup/iplookup"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/geoiplookup/resolverlookup"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/resources"
)

// DataUsageFunc provides information about data usage.
type DataUsageFunc func(dloadKiB, uploadKiB float64)

// ProgressFunc provides information about an experiment progress.
type ProgressFunc func(percentage float64, message string)

// Session contains information on a measurement session.
type Session struct {
	// AvailableBouncers contains the available bouncers.
	AvailableBouncers []model.Service

	// AvailableCollectors contains the available collectors.
	AvailableCollectors []model.Service

	// AvailableTestHelpers contains the available test helpers.
	AvailableTestHelpers map[string][]model.Service

	// DataUsage is called to provide updates on data usage.
	DataUsage DataUsageFunc

	// HTTPDefaultClient is the default HTTP client to use.
	HTTPDefaultClient *http.Client

	// HTTPNoProxyClient is a non-proxied HTTP client.
	HTTPNoProxyClient *http.Client

	// Logger is the log emitter.
	Logger log.Logger

	// Location is the probe location.
	Location *model.LocationInfo

	// Progress is called to provide information about progress.
	Progress ProgressFunc

	// SoftwareName contains the software name.
	SoftwareName string

	// SoftwareVersion contains the software version.
	SoftwareVersion string

	// WorkDir is the session's working directory.
	WorkDir string
}

// New creates a new experiments session.
func New(logger log.Logger, softwareName, softwareVersion string) *Session {
	return &Session{
		DataUsage: func(dloadKiB, uploadKiB float64) {
			logger.Infof("data usage: %f/%f down/up KiB", dloadKiB, uploadKiB)
		},
		HTTPDefaultClient: httpx.NewTracingProxyingClient(
			logger, http.ProxyFromEnvironment,
		),
		HTTPNoProxyClient: httpx.NewTracingProxyingClient(logger, nil),
		Logger:            logger,
		Progress: func(percentage float64, message string) {
			logger.Infof("[%4.1f%%] %s", percentage*100, message)
		},
		SoftwareName:    softwareName,
		SoftwareVersion: softwareVersion,
		WorkDir:         os.TempDir(),
	}
}

// ProbeASNString returns the probe ASN as a string.
func (s *Session) ProbeASNString() string {
	asn := model.DefaultProbeASN
	if s.Location != nil {
		asn = s.Location.ASN
	}
	return fmt.Sprintf("AS%d", asn)
}

// ProbeCC returns the probe CC.
func (s *Session) ProbeCC() string {
	cc := model.DefaultProbeCC
	if s.Location != nil {
		cc = s.Location.CountryCode
	}
	return cc
}

// ProbeNetworkName returns the probe network name.
func (s *Session) ProbeNetworkName() string {
	nn := model.DefaultProbeNetworkName
	if s.Location != nil {
		nn = s.Location.NetworkName
	}
	return nn
}

// ProbeIP returns the probe IP.
func (s *Session) ProbeIP() string {
	ip := model.DefaultProbeIP
	if s.Location != nil {
		ip = s.Location.ProbeIP
	}
	return ip
}

func (s *Session) fetchResourcesIdempotent(ctx context.Context) error {
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
func (s *Session) CABundlePath() string {
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

func (s *Session) lookupCollectors(ctx context.Context) error {
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableCollectors, err = client.GetCollectors(ctx)
		return
	})
}

func (s *Session) lookupTestHelpers(ctx context.Context) error {
	return s.queryBouncer(ctx, func(client *bouncer.Client) (err error) {
		s.AvailableTestHelpers, err = client.GetTestHelpers(ctx)
		return
	})
}

// LookupBackends discovers the available OONI backends.
func (s *Session) LookupBackends(ctx context.Context) (err error) {
	err = s.lookupCollectors(ctx)
	if err != nil {
		return
	}
	err = s.lookupTestHelpers(ctx)
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
	addrs, err := resolverlookup.Do(ctx, nil)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return "", errors.New("No resolver IPs returned")
	}
	return addrs[0], nil
}

// LookupLocation discovers details on the probe location.
func (s *Session) LookupLocation(ctx context.Context) error {
	if s.Location != nil {
		return nil
	}
	err := s.fetchResourcesIdempotent(ctx)
	if err != nil {
		return err
	}
	probeIP, err := s.lookupProbeIP(ctx)
	if err != nil {
		return err
	}
	asn, org, err := s.lookupProbeNetwork(s.ASNDatabasePath(), probeIP)
	if err != nil {
		return err
	}
	cc, err := s.lookupProbeCC(s.CountryDatabasePath(), probeIP)
	if err != nil {
		return err
	}
	resolverIP, err := s.lookupResolverIP(ctx)
	if err != nil {
		return err
	}
	s.Location = &model.LocationInfo{
		ASN:         asn,
		CountryCode: cc,
		NetworkName: org,
		ProbeIP:     probeIP,
		ResolverIP:  resolverIP,
	}
	return nil
}
