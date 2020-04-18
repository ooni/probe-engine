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
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/ClientLibrary/clientlib"
	"github.com/ooni/probe-engine/experiment/httpheader"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	testName    = "psiphon"
	testVersion = "0.3.2"
)

// Config contains the experiment's configuration.
type Config struct {
	// WorkDir is the directory where Psiphon should store
	// its configuration database.
	WorkDir string `ooni:"experiment working directory"`
}

// TestKeys contains the experiment's result.
//
// This is what will end up into the Measurement.TestKeys field
// when you run this experiment.
type TestKeys struct {
	Agent         string                     `json:"agent"`
	BootstrapTime float64                    `json:"bootstrap_time"`
	Failure       *string                    `json:"failure"`
	MaxRuntime    float64                    `json:"max_runtime"`
	NetworkEvents archival.NetworkEventsList `json:"network_events"`
	Queries       archival.DNSQueriesList    `json:"queries"`
	Requests      archival.RequestList       `json:"requests"`
	SOCKSProxy    string                     `json:"socksproxy"`
	TLSHandshakes archival.TLSHandshakesList `json:"tls_handshakes"`
}

func registerExtensions(m *model.Measurement) {
	archival.ExtHTTP.AddTo(m)
	archival.ExtDNS.AddTo(m)
	archival.ExtNetevents.AddTo(m)
	archival.ExtTLSHandshake.AddTo(m)
}

type runner struct {
	beginning      time.Time
	callbacks      model.ExperimentCallbacks
	config         Config
	ioutilReadFile func(filename string) ([]byte, error)
	osMkdirAll     func(path string, perm os.FileMode) error
	osRemoveAll    func(path string) error
	testkeys       *TestKeys
}

func newRunner(
	config Config, callbacks model.ExperimentCallbacks,
	beginning time.Time,
) *runner {
	return &runner{
		beginning:      beginning,
		callbacks:      callbacks,
		config:         config,
		ioutilReadFile: ioutil.ReadFile,
		osMkdirAll:     os.MkdirAll,
		osRemoveAll:    os.RemoveAll,
		testkeys:       new(TestKeys),
	}
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
	ctx context.Context, port int, logger model.Logger,
) error {
	// TODO(bassosimone): understand if there is a way to ask
	// the tunnel the number of bytes sent and received
	r.testkeys.Agent = "redirect"
	r.testkeys.SOCKSProxy = fmt.Sprintf("127.0.0.1:%d", port)
	saver := new(trace.Saver)
	clnt := &http.Client{Transport: httptransport.New(httptransport.Config{
		ContextByteCounting: true,
		Logger:              logger,
		SOCKS5Proxy:         r.testkeys.SOCKSProxy,
		Saver:               saver,
	})}
	defer clnt.CloseIdleConnections()
	req, err := http.NewRequest("GET", "https://www.google.com/humans.txt", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", httpheader.RandomAccept())
	req.Header.Set("Accept-Language", httpheader.RandomAcceptLanguage())
	req.Header.Set("User-Agent", httpheader.RandomUserAgent())
	begin := time.Now()
	defer func() {
		events := saver.Read()
		r.testkeys.Queries = append(
			r.testkeys.Queries, archival.NewDNSQueriesList(begin, events)...,
		)
		r.testkeys.NetworkEvents = append(
			r.testkeys.NetworkEvents, archival.NewNetworkEventsList(begin, events)...,
		)
		r.testkeys.Requests = append(
			r.testkeys.Requests, archival.NewRequestList(begin, events)...,
		)
		r.testkeys.TLSHandshakes = append(
			r.testkeys.TLSHandshakes, archival.NewTLSHandshakesList(begin, events)...,
		)
	}()
	resp, err := clnt.Do(req.WithContext(ctx))
	if err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	if err := resp.Body.Close(); err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	return nil
}

func (r *runner) run(
	ctx context.Context,
	logger model.Logger,
	fetchPsiphonConfig func(ctx context.Context) ([]byte, error),
) error {
	configJSON, err := fetchPsiphonConfig(ctx)
	if err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	workdir, err := r.makeworkingdir()
	if err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	start := time.Now()
	tunnel, err := r.starttunnel(ctx, configJSON, clientlib.Parameters{
		DataRootDirectory: &workdir,
	})
	if err != nil {
		s := err.Error()
		r.testkeys.Failure = &s
		return err
	}
	r.testkeys.BootstrapTime = time.Since(start).Seconds()
	defer tunnel.Stop()
	return r.usetunnel(ctx, tunnel.SOCKSProxyPort, logger)
}

type measurer struct {
	config Config
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) printprogress(
	ctx context.Context, wg *sync.WaitGroup,
	maxruntime int, callbacks model.ExperimentCallbacks,
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

func (m *measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	clnt, err := sess.NewOrchestraClient(ctx)
	if err != nil {
		return err
	}
	const maxruntime = 60
	ctx, cancel := context.WithTimeout(ctx, maxruntime*time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go m.printprogress(ctx, &wg, maxruntime, callbacks)
	r := newRunner(m.config, callbacks, measurement.MeasurementStartTimeSaved)
	registerExtensions(measurement)
	measurement.TestKeys = r.testkeys
	r.testkeys.MaxRuntime = maxruntime
	err = r.run(ctx, sess.Logger(), clnt.FetchPsiphonConfig)
	cancel()
	wg.Wait()
	return err
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}
