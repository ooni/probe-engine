// Package mockable contains mockable objects
package mockable

import (
	"context"
	"net/http"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/internal/orchestra/testorchestra"
	"github.com/ooni/probe-engine/model"
)

// ExperimentSession is a mockable ExperimentSession.
type ExperimentSession struct {
	MockableASNDatabasePath      string
	MockableCABundlePath         string
	MockableExplicitProxy        bool
	MockableTestHelpers          map[string][]model.Service
	MockableHTTPClient           *http.Client
	MockableLogger               model.Logger
	MockableOrchestraClient      model.ExperimentOrchestraClient
	MockableOrchestraClientError error
	MockableProbeASNString       string
	MockableProbeCC              string
	MockableProbeIP              string
	MockableProbeNetworkName     string
	MockableSoftwareName         string
	MockableSoftwareVersion      string
	MockableTempDir              string
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

// ExplicitProxy implements ExperimentSession.ExplicitProxy
func (sess *ExperimentSession) ExplicitProxy() bool {
	return sess.MockableExplicitProxy
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

// Logger implements ExperimentSession.Logger
func (sess *ExperimentSession) Logger() model.Logger {
	return sess.MockableLogger
}

// NewOrchestraClient implements ExperimentSession.NewOrchestraClient
func (sess *ExperimentSession) NewOrchestraClient(ctx context.Context) (model.ExperimentOrchestraClient, error) {
	if sess.MockableOrchestraClient != nil {
		return sess.MockableOrchestraClient, nil
	}
	if sess.MockableOrchestraClientError != nil {
		return nil, sess.MockableOrchestraClientError
	}
	clnt := orchestra.NewClient(
		http.DefaultClient,
		log.Log,
		"miniooni/0.1.0-dev",
		statefile.New(kvstore.NewMemoryKeyValueStore()),
	)
	clnt.OrchestrateBaseURL = "https://ps-test.ooni.io"
	clnt.RegistryBaseURL = "https://ps-test.ooni.io"
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

// UserAgent implements ExperimentSession.UserAgent
func (sess *ExperimentSession) UserAgent() string {
	return sess.MockableUserAgent
}
