// Package psiphon implements the psiphon network experiment.
package psiphon

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/ClientLibrary/clientlib"
	"github.com/ooni/probe-engine/httpx/fetch"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "psiphon"
	testVersion = "0.3.0"
)

// NewReporter creates a new reporter.
func NewReporter(cs *session.Session) *experiment.Reporter {
	return experiment.NewReporter(cs, testName, testVersion)
}

// Config contains the experiment's configuration.
type Config struct {
	// ConfigFilePath is the path where Psiphon config file is located.
	ConfigFilePath string

	// Logger is the logger to use
	Logger log.Logger

	// UserAgent is the user agent to use
	UserAgent string

	// WorkDir is the directory where Psiphon should store
	// its configuration database.
	WorkDir string
}

// TestKeys contains the experiment's result.
//
// This is what will end up into the Measurement.TestKeys field
// when you run this experiment.
type TestKeys struct {
	// Failure contains the failure that occurred.
	Failure string `json:"failure"`

	// BootstrapTime is the time it took to bootstrap Psiphon.
	BootstrapTime float64 `json:"bootstrap_time"`
}

// osRemoveAll is a mockable os.RemoveAll
var osRemoveAll = os.RemoveAll

// osMkdirAll is a mockable os.MkdirAll
var osMkdirAll = os.MkdirAll

// ioutilReadFile is a mockable ioutil.ReadFile
var ioutilReadFile = ioutil.ReadFile

func processconfig(config Config) ([]byte, clientlib.Parameters, error) {
	if config.WorkDir == "" {
		return nil, clientlib.Parameters{}, errors.New("WorkDir is empty")
	}
	const testdirname = "oonipsiphon"
	workdir := filepath.Join(config.WorkDir, testdirname)
	err := osRemoveAll(workdir)
	if err != nil {
		return nil, clientlib.Parameters{}, err
	}
	err = osMkdirAll(workdir, 0700)
	if err != nil {
		return nil, clientlib.Parameters{}, err
	}
	params := clientlib.Parameters{
		DataRootDirectory: &workdir,
	}
	configJSON, err := ioutilReadFile(config.ConfigFilePath)
	if err != nil {
		return nil, clientlib.Parameters{}, err
	}
	return configJSON, params, nil
}

// usetunnel is a mockable function that uses the tunnel
var usetunnel = func(
	ctx context.Context, t *clientlib.PsiphonTunnel, config Config,
) error {
	_, err := (&fetch.Client{
		HTTPClient: httpx.NewTracingProxyingClient(
			config.Logger,
			func(req *http.Request) (*url.URL, error) {
				return &url.URL{
					Scheme: "socks5",
					Host:   fmt.Sprintf("127.0.0.1:%d", t.SOCKSProxyPort),
				}, nil
			},
		),
		Logger:    config.Logger,
		UserAgent: config.UserAgent,
	}).Fetch(ctx, "https://www.google.com/humans.txt")
	return err
}

// clientlibStartTunnel is a mockable clientlib.StartTunnel
var clientlibStartTunnel = clientlib.StartTunnel

// Run runs a psiphon experiment
func Run(
	ctx context.Context,
	measurement *model.Measurement,
	config Config,
) error {
	testkeys := &TestKeys{}
	measurement.TestKeys = testkeys
	configJSON, params, err := processconfig(config)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	start := time.Now()
	tunnel, err := clientlibStartTunnel(ctx, configJSON, "", params, nil, nil)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	testkeys.BootstrapTime = time.Now().Sub(start).Seconds()
	defer tunnel.Stop()
	err = usetunnel(ctx, tunnel, config)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	return nil
}
