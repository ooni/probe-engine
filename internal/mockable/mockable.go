// Package mockable contains mockable objects
package mockable

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/runtimex"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
	"github.com/ooni/probe-engine/probeservices/testorchestra"
)

// ExperimentSession is a mockable ExperimentSession.
type ExperimentSession struct {
	MockableASNDatabasePath      string
	MockableCABundlePath         string
	MockableTestHelpers          map[string][]model.Service
	MockableHTTPClient           *http.Client
	MockableLogger               model.Logger
	MockableMaybeStartTunnelErr  error
	MockableOrchestraClient      model.ExperimentOrchestraClient
	MockableOrchestraClientError error
	MockableProbeASNString       string
	MockableProbeCC              string
	MockableProbeIP              string
	MockableProbeNetworkName     string
	MockableProxyURL             *url.URL
	MockableResolverIP           string
	MockableSoftwareName         string
	MockableSoftwareVersion      string
	MockableTempDir              string
	MockableTorArgs              []string
	MockableTorBinary            string
	MockableTunnelBootstrapTime  time.Duration
	MockableUserAgent            string
}

// ASNDatabasePath implements ExperimentSession.ASNDatabasePath
func (sess *ExperimentSession) ASNDatabasePath() string {
	return sess.MockableASNDatabasePath
}

// CABundlePath implements ExperimentSession.CABundlePath
func (sess *ExperimentSession) CABundlePath() string {
	return sess.MockableCABundlePath
}

// GetTestHelpersByName implements ExperimentSession.GetTestHelpersByName
func (sess *ExperimentSession) GetTestHelpersByName(name string) ([]model.Service, bool) {
	services, okay := sess.MockableTestHelpers[name]
	return services, okay
}

// DefaultHTTPClient implements ExperimentSession.DefaultHTTPClient
func (sess *ExperimentSession) DefaultHTTPClient() *http.Client {
	return sess.MockableHTTPClient
}

// KeyValueStore returns the configured key-value store.
func (sess *ExperimentSession) KeyValueStore() model.KeyValueStore {
	return kvstore.NewMemoryKeyValueStore()
}

// Logger implements ExperimentSession.Logger
func (sess *ExperimentSession) Logger() model.Logger {
	return sess.MockableLogger
}

// MaybeStartTunnel implements ExperimentSession.MaybeStartTunnel
func (sess *ExperimentSession) MaybeStartTunnel(ctx context.Context, name string) error {
	return sess.MockableMaybeStartTunnelErr
}

// NewOrchestraClient implements ExperimentSession.NewOrchestraClient
func (sess *ExperimentSession) NewOrchestraClient(ctx context.Context) (model.ExperimentOrchestraClient, error) {
	if sess.MockableOrchestraClient != nil {
		return sess.MockableOrchestraClient, nil
	}
	if sess.MockableOrchestraClientError != nil {
		return nil, sess.MockableOrchestraClientError
	}
	clnt, err := probeservices.NewClient(sess, model.Service{
		Address: "https://ams-pg.ooni.org/",
		Type:    "https",
	})
	runtimex.PanicOnError(err, "orchestra.NewClient should not fail here")
	meta := testorchestra.MetadataFixture()
	if err := clnt.MaybeRegister(ctx, meta); err != nil {
		return nil, err
	}
	if err := clnt.MaybeLogin(ctx); err != nil {
		return nil, err
	}
	return clnt, nil
}

// ProbeASNString implements ExperimentSession.ProbeASNString
func (sess *ExperimentSession) ProbeASNString() string {
	return sess.MockableProbeASNString
}

// ProbeCC implements ExperimentSession.ProbeCC
func (sess *ExperimentSession) ProbeCC() string {
	return sess.MockableProbeCC
}

// ProbeIP implements ExperimentSession.ProbeIP
func (sess *ExperimentSession) ProbeIP() string {
	return sess.MockableProbeIP
}

// ProbeNetworkName implements ExperimentSession.ProbeNetworkName
func (sess *ExperimentSession) ProbeNetworkName() string {
	return sess.MockableProbeNetworkName
}

// ProxyURL implements ExperimentSession.ProxyURL
func (sess *ExperimentSession) ProxyURL() *url.URL {
	return sess.MockableProxyURL
}

// ResolverIP implements ExperimentSession.ResolverIP
func (sess *ExperimentSession) ResolverIP() string {
	return sess.MockableResolverIP
}

// SoftwareName implements ExperimentSession.SoftwareName
func (sess *ExperimentSession) SoftwareName() string {
	return sess.MockableSoftwareName
}

// SoftwareVersion implements ExperimentSession.SoftwareVersion
func (sess *ExperimentSession) SoftwareVersion() string {
	return sess.MockableSoftwareVersion
}

// TempDir implements ExperimentSession.TempDir
func (sess *ExperimentSession) TempDir() string {
	return sess.MockableTempDir
}

// TorArgs implements ExperimentSession.TorArgs.
func (sess *ExperimentSession) TorArgs() []string {
	return sess.MockableTorArgs
}

// TorBinary implements ExperimentSession.TorBinary.
func (sess *ExperimentSession) TorBinary() string {
	return sess.MockableTorBinary
}

// TunnelBootstrapTime implements ExperimentSession.TunnelBootstrapTime
func (sess *ExperimentSession) TunnelBootstrapTime() time.Duration {
	return sess.MockableTunnelBootstrapTime
}

// UserAgent implements ExperimentSession.UserAgent
func (sess *ExperimentSession) UserAgent() string {
	return sess.MockableUserAgent
}

var _ model.ExperimentSession = &ExperimentSession{}

// ExperimentOrchestraClient is the experiment's view of
// a client for querying the OONI orchestra.
type ExperimentOrchestraClient struct {
	MockableFetchPsiphonConfigResult []byte
	MockableFetchPsiphonConfigErr    error
	MockableFetchTorTargetsResult    map[string]model.TorTarget
	MockableFetchTorTargetsErr       error
	MockableFetchURLListResult       []model.URLInfo
	MockableFetchURLListErr          error
}

// FetchPsiphonConfig implements ExperimentOrchestraClient.FetchPsiphonConfig
func (c ExperimentOrchestraClient) FetchPsiphonConfig(
	ctx context.Context) ([]byte, error) {
	return c.MockableFetchPsiphonConfigResult, c.MockableFetchPsiphonConfigErr
}

// FetchTorTargets implements ExperimentOrchestraClient.TorTargets
func (c ExperimentOrchestraClient) FetchTorTargets(
	ctx context.Context, cc string) (map[string]model.TorTarget, error) {
	return c.MockableFetchTorTargetsResult, c.MockableFetchTorTargetsErr
}

// FetchURLList implements ExperimentOrchestraClient.FetchURLList.
func (c ExperimentOrchestraClient) FetchURLList(
	ctx context.Context, config model.URLListConfig) ([]model.URLInfo, error) {
	return c.MockableFetchURLListResult, c.MockableFetchURLListErr
}

var _ model.ExperimentOrchestraClient = ExperimentOrchestraClient{}
