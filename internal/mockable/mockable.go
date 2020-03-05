// Package mockable contains mockable objects
package mockable

import (
	"context"
	"errors"
	"net/http"

	"github.com/ooni/probe-engine/model"
)

// ExperimentSession is a mockable ExperimentSession.
type ExperimentSession struct {
	MockableASNDatabasePath  string
	MockableCABundlePath     string
	MockableExplicitProxy    bool
	MockableTestHelpers      map[string][]model.Service
	MockableHTTPClient       *http.Client
	MockableLogger           model.Logger
	MockableOrchestraClient  model.ExperimentOrchestraClient
	MockableProbeASNString   string
	MockableProbeCC          string
	MockableProbeIP          string
	MockableProbeNetworkName string
	MockableSoftwareName     string
	MockableSoftwareVersion  string
	MockableTempDir          string
	MockableUserAgent        string
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
	return nil, errors.New("no mockable orchestra client")
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
