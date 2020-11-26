package model

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/internal/humanizex"
)

// ExperimentOrchestraClient is the experiment's view of
// a client for querying the OONI orchestra API.
type ExperimentOrchestraClient interface {
	FetchPsiphonConfig(ctx context.Context) ([]byte, error)
	FetchTorTargets(ctx context.Context, cc string) (map[string]TorTarget, error)
	FetchURLList(ctx context.Context, config URLListConfig) ([]URLInfo, error)
}

// ExperimentSession is the experiment's view of a session.
type ExperimentSession interface {
	ASNDatabasePath() string
	GetTestHelpersByName(name string) ([]Service, bool)
	DefaultHTTPClient() *http.Client
	Logger() Logger
	MaybeStartTunnel(ctx context.Context, name string) error
	NewOrchestraClient(ctx context.Context) (ExperimentOrchestraClient, error)
	ProbeCC() string
	ProxyURL() *url.URL
	MaybeResolverIP() string
	TempDir() string
	TunnelBootstrapTime() time.Duration
	UserAgent() string
}

// ExperimentCallbacks contains experiment event-handling callbacks
type ExperimentCallbacks interface {
	// OnDataUsage provides information about data usage.
	//
	// This callback is deprecated and will be removed once we have
	// removed the dependency on Measurement Kit.
	OnDataUsage(dloadKiB, uploadKiB float64)

	// OnProgress provides information about an experiment progress.
	OnProgress(percentage float64, message string)
}

// PrinterCallbacks is the default event handler
type PrinterCallbacks struct {
	Logger
}

// NewPrinterCallbacks returns a new default callback handler
func NewPrinterCallbacks(logger Logger) PrinterCallbacks {
	return PrinterCallbacks{Logger: logger}
}

// OnDataUsage provides information about data usage.
func (d PrinterCallbacks) OnDataUsage(dloadKiB, uploadKiB float64) {
	d.Logger.Infof("experiment: recv %s, sent %s",
		humanizex.SI(dloadKiB*1024, "byte"),
		humanizex.SI(uploadKiB*1024, "byte"),
	)
}

// OnProgress provides information about an experiment progress.
func (d PrinterCallbacks) OnProgress(percentage float64, message string) {
	d.Logger.Infof("[%5.1f%%] %s", percentage*100, message)
}

// ExperimentMeasurer is the interface that allows to run a
// measurement for a specific experiment.
type ExperimentMeasurer interface {
	// ExperimentName returns the experiment name.
	ExperimentName() string

	// ExperimentVersion returns the experiment version.
	ExperimentVersion() string

	// Run runs the experiment with the specified context, session,
	// measurement, and experiment calbacks. This method should only
	// return an error in case the experiment could not run (e.g.,
	// a required input is missing). Otherwise, the code should just
	// set the relevant OONI error inside of the measurmeent and
	// return nil. This is important because the caller may not submit
	// the measurement if this method returns an error.
	Run(
		ctx context.Context, sess ExperimentSession,
		measurement *Measurement, callbacks ExperimentCallbacks,
	) error

	// GetSummaryKeys returns summary keys expected by ooni/probe-cli.
	GetSummaryKeys(*Measurement) (interface{}, error)
}
