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
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/httpheader"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
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
	Agent         string                          `json:"agent"`
	BootstrapTime float64                         `json:"bootstrap_time"`
	Failure       *string                         `json:"failure"`
	MaxRuntime    float64                         `json:"max_runtime"`
	Queries       oonidatamodel.DNSQueriesList    `json:"queries"`
	Requests      oonidatamodel.RequestList       `json:"requests"`
	SOCKSProxy    string                          `json:"socksproxy"`
	TLSHandshakes oonidatamodel.TLSHandshakesList `json:"tls_handshakes"`
}

type runner struct {
	beginning      time.Time
	callbacks      handler.Callbacks
	config         Config
	ioutilReadFile func(filename string) ([]byte, error)
	osMkdirAll     func(path string, perm os.FileMode) error
	osRemoveAll    func(path string) error
	testkeys       *TestKeys
}

func newRunner(
	config Config, callbacks handler.Callbacks,
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
	ctx context.Context, port int, logger log.Logger,
) error {
	r.testkeys.Agent = "redirect"
	r.testkeys.SOCKSProxy = fmt.Sprintf("127.0.0.1:%d", port)
	results := oonitemplates.HTTPDo(ctx, oonitemplates.HTTPDoConfig{
		Accept:         httpheader.RandomAccept(),
		AcceptLanguage: httpheader.RandomAcceptLanguage(),
		Beginning:      r.beginning,
		Handler:        netxlogger.NewHandler(logger),
		Method:         "GET",
		ProxyFunc: func(req *http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "socks5",
				Host:   r.testkeys.SOCKSProxy,
			}, nil
		},
		URL:       "https://www.google.com/humans.txt",
		UserAgent: httpheader.RandomUserAgent(),
	})
	r.testkeys.Queries = append(
		r.testkeys.Queries, oonidatamodel.NewDNSQueriesList(results.TestKeys)...,
	)
	r.testkeys.Requests = append(
		r.testkeys.Requests, oonidatamodel.NewRequestList(results.TestKeys)...,
	)
	r.testkeys.TLSHandshakes = append(
		r.testkeys.TLSHandshakes, oonidatamodel.NewTLSHandshakesList(results.TestKeys)...,
	)
	// TODO(bassosimone): understand if there is a way to ask
	// the tunnel the number of bytes sent and received
	receivedBytes := results.TestKeys.ReceivedBytes
	sentBytes := results.TestKeys.SentBytes
	r.callbacks.OnDataUsage(
		float64(receivedBytes)/1024.0, // downloaded
		float64(sentBytes)/1024.0,     // uploaded
	)
	if results.Error != nil {
		s := results.Error.Error()
		r.testkeys.Failure = &s
		return results.Error
	}
	return nil
}

func (r *runner) run(
	ctx context.Context,
	logger log.Logger,
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
	measurement.TestKeys = r.testkeys
	r.testkeys.MaxRuntime = maxruntime
	err = r.run(ctx, sess.Logger, clnt.FetchPsiphonConfig)
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
