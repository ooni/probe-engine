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
	"sync"
	"time"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/ClientLibrary/clientlib"
	netxlogger "github.com/ooni/netx/x/logger"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/useragent"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// TestKeys contains the experiment's result.
//
// This is what will end up into the Measurement.TestKeys field
// when you run this experiment.
type TestKeys struct {
	// BootstrapTime is the time it took to bootstrap Psiphon.
	BootstrapTime float64 `json:"bootstrap_time"`

	// Failure contains the failure that occurred.
	Failure string `json:"failure"`
}

type runner struct {
	callbacks      handler.Callbacks
	config         Config
	ioutilReadFile func(filename string) ([]byte, error)
	osMkdirAll     func(path string, perm os.FileMode) error
	osRemoveAll    func(path string) error
	testkeys       *TestKeys
}

func newRunner(config Config, callbacks handler.Callbacks) *runner {
	return &runner{
		callbacks:      callbacks,
		config:         config,
		ioutilReadFile: ioutil.ReadFile,
		osMkdirAll:     os.MkdirAll,
		osRemoveAll:    os.RemoveAll,
		testkeys:       new(TestKeys),
	}
}

func (r *runner) readconfig() ([]byte, error) {
	configJSON, err := r.ioutilReadFile(r.config.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	return configJSON, nil
}

func (r *runner) makeworkingdir() (string, error) {
	if r.config.WorkDir == "" {
		return "", errors.New("WorkDir is empty")
	}
	const testdirname = "oonipsiphon"
	workdir := filepath.Join(r.config.WorkDir, testdirname)
	if err := r.osRemoveAll(workdir); err != nil {
		return "", err
	}
	if err := r.osMkdirAll(workdir, 0700); err != nil {
		return "", err
	}
	return workdir, nil
}

func (r *runner) starttunnel(
	ctx context.Context, configJSON []byte,
	params clientlib.Parameters,
) (*clientlib.PsiphonTunnel, error) {
	return clientlib.StartTunnel(ctx, configJSON, "", params, nil, nil)
}

func (r *runner) usetunnel(
	ctx context.Context, port int, logger log.Logger,
) error {
	// TODO(bassosimone): here we should store the results of
	// fetching the page using psiphon and http.
	results := porcelain.HTTPDo(ctx, porcelain.HTTPDoConfig{
		Handler: netxlogger.NewHandler(logger),
		Method:  "GET",
		ProxyFunc: func(req *http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "socks5",
				Host:   fmt.Sprintf("127.0.0.1:%d", port),
			}, nil
		},
		URL:       "https://www.google.com/humans.txt",
		UserAgent: useragent.Random(),
	})
	// TODO(bassosimone): understand if there is a way to ask
	// the tunnel the number of bytes sent and/or received
	receivedBytes := results.TestKeys.ReceivedBytes
	sentBytes := results.TestKeys.SentBytes
	r.callbacks.OnDataUsage(
		float64(receivedBytes)/1024.0, // downloaded
		float64(sentBytes)/1024.0,     // uploaded
	)
	if results.Error != nil {
		r.testkeys.Failure = results.Error.Error()
		return results.Error
	}
	return nil
}

func (r *runner) run(ctx context.Context, logger log.Logger) error {
	configJSON, err := r.readconfig()
	if err != nil {
		r.testkeys.Failure = err.Error()
		return err
	}
	workdir, err := r.makeworkingdir()
	if err != nil {
		r.testkeys.Failure = err.Error()
		return err
	}
	start := time.Now()
	tunnel, err := r.starttunnel(ctx, configJSON, clientlib.Parameters{
		DataRootDirectory: &workdir,
	})
	if err != nil {
		r.testkeys.Failure = err.Error()
		return err
	}
	r.testkeys.BootstrapTime = time.Since(start).Seconds()
	defer tunnel.Stop()
	return r.usetunnel(ctx, tunnel.SOCKSProxyPort, logger)
}

type measurer struct {
	config Config
}

func (m *measurer) printprogress(
	ctx context.Context, wg *sync.WaitGroup,
	maxruntime int, callbacks handler.Callbacks,
) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	step := 1 / float64(maxruntime)
	var progress float64
	defer callbacks.OnProgress(1.0, "psiphon experiment complete")
	defer wg.Done()
	for {
		select {
		case <-ticker.C:
			progress += step
			callbacks.OnProgress(progress, "psiphon experiment running")
		case <-ctx.Done():
			return
		}
	}
}

func (m *measurer) measure(
	ctx context.Context, sess *session.Session,
	measurement *model.Measurement, callbacks handler.Callbacks,
) error {
	const maxruntime = 30
	ctx, cancel := context.WithTimeout(ctx, maxruntime*time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go m.printprogress(ctx, &wg, maxruntime, callbacks)
	r := newRunner(m.config, callbacks)
	measurement.TestKeys = r.testkeys
	err := r.run(ctx, sess.Logger)
	cancel()
	wg.Wait()
	return err
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	m := &measurer{config: config}
	return experiment.New(sess, testName, testVersion, m.measure)
}
