// +build !nopsiphon

// Package psiphon implements the psiphon network experiment. This
// implements, in particular, v0.2.0 of the spec.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-015-psiphon.md
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
	netxlogger "github.com/ooni/netx/x/logger"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/useragent"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

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
	sess *session.Session,
) error {
	_, err := porcelain.HTTPDo(ctx, porcelain.HTTPDoConfig{
		Handler: netxlogger.NewHandler(sess.Logger),
		Method:  "GET",
		ProxyFunc: func(req *http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "socks5",
				Host:   fmt.Sprintf("127.0.0.1:%d", t.SOCKSProxyPort),
			}, nil
		},
		URL:       "https://www.google.com/humans.txt",
		UserAgent: useragent.Random(),
	})
	return err
}

// clientlibStartTunnel is a mockable clientlib.StartTunnel
var clientlibStartTunnel = clientlib.StartTunnel

func run(
	ctx context.Context,
	sess *session.Session,
	measurement *model.Measurement,
	config Config,
	callbacks handler.Callbacks,
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
	err = usetunnel(ctx, tunnel, config, sess)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(
		sess, testName, testVersion,
		func(c context.Context, s *session.Session, m *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return run(c, s, m, config, callbacks)
		})
}
