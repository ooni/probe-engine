// Package tor contains the tor experiment.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-023-tor.md
package tor

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/httpheader"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	// parallelism is the number of parallel threads we use for this experiment
	parallelism = 2

	// testName is the name of this experiment
	testName = "tor"

	// testVersion is th version of this experiment
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct{}

// Summary contains a summary of what happened.
type Summary struct {
	Failure *string `json:"failure"`
}

// TargetResults contains the results of measuring a target.
type TargetResults struct {
	Agent          string                          `json:"agent"`
	Failure        *string                         `json:"failure"`
	NetworkEvents  oonidatamodel.NetworkEventsList `json:"network_events"`
	Queries        oonidatamodel.DNSQueriesList    `json:"queries"`
	Requests       oonidatamodel.RequestList       `json:"requests"`
	Summary        map[string]Summary              `json:"summary"`
	TargetAddress  string                          `json:"target_address"`
	TargetName     string                          `json:"target_name,omitempty"`
	TargetProtocol string                          `json:"target_protocol"`
	TCPConnect     oonidatamodel.TCPConnectList    `json:"tcp_connect"`
	TLSHandshakes  oonidatamodel.TLSHandshakesList `json:"tls_handshakes"`
}

func registerExtensions(m *model.Measurement) {
	oonidatamodel.ExtHTTP.AddTo(m)
	oonidatamodel.ExtNetevents.AddTo(m)
	oonidatamodel.ExtDNS.AddTo(m)
	oonidatamodel.ExtTCPConnect.AddTo(m)
	oonidatamodel.ExtTLSHandshake.AddTo(m)
}

// fillSummary fills the Summary field used by the UI.
func (tr *TargetResults) fillSummary() {
	tr.Summary = make(map[string]Summary)
	if len(tr.TCPConnect) < 1 {
		return
	}
	tr.Summary[modelx.ConnectOperation] = Summary{
		Failure: tr.TCPConnect[0].Status.Failure,
	}
	switch tr.TargetProtocol {
	case "dir_port":
		// The UI currently doesn't care about this protocol
		// as long as drawing a table is concerned.
	case "obfs4":
		// We currently only perform an OBFS4 handshake, hence
		// the final Failure is the handshake result
		tr.Summary["handshake"] = Summary{
			Failure: tr.Failure,
		}
	case "or_port_dirauth", "or_port":
		if len(tr.TLSHandshakes) < 1 {
			return
		}
		tr.Summary["handshake"] = Summary{
			Failure: tr.TLSHandshakes[0].Failure,
		}
	}
}

// TestKeys contains tor test keys.
type TestKeys struct {
	DirPortTotal            int64                    `json:"dir_port_total"`
	DirPortAccessible       int64                    `json:"dir_port_accessible"`
	OBFS4Total              int64                    `json:"obfs4_total"`
	OBFS4Accessible         int64                    `json:"obfs4_accessible"`
	ORPortDirauthTotal      int64                    `json:"or_port_dirauth_total"`
	ORPortDirauthAccessible int64                    `json:"or_port_dirauth_accessible"`
	ORPortTotal             int64                    `json:"or_port_total"`
	ORPortAccessible        int64                    `json:"or_port_accessible"`
	Targets                 map[string]TargetResults `json:"targets"`
}

func (tk *TestKeys) fillToplevelKeys() {
	for _, value := range tk.Targets {
		switch value.TargetProtocol {
		case "dir_port":
			tk.DirPortTotal++
			if value.Failure == nil {
				tk.DirPortAccessible++
			}
		case "obfs4":
			tk.OBFS4Total++
			if value.Failure == nil {
				tk.OBFS4Accessible++
			}
		case "or_port_dirauth":
			tk.ORPortDirauthTotal++
			if value.Failure == nil {
				tk.ORPortDirauthAccessible++
			}
		case "or_port":
			tk.ORPortTotal++
			if value.Failure == nil {
				tk.ORPortAccessible++
			}
		}
	}
}

type measurer struct {
	config             Config
	fetchTorTargets    func(ctx context.Context, clnt model.ExperimentOrchestraClient) (map[string]model.TorTarget, error)
	newOrchestraClient func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error)
}

func newMeasurer(config Config) *measurer {
	return &measurer{
		config: config,
		fetchTorTargets: func(ctx context.Context, clnt model.ExperimentOrchestraClient) (map[string]model.TorTarget, error) {
			return clnt.FetchTorTargets(ctx)
		},
		newOrchestraClient: func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error) {
			return sess.NewOrchestraClient(ctx)
		},
	}
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) error {
	targets, err := m.gimmeTargets(ctx, sess)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(
		ctx, 15*time.Second*time.Duration(len(targets)),
	)
	defer cancel()
	registerExtensions(measurement)
	m.measureTargets(ctx, sess, measurement, callbacks, targets)
	return nil
}

func (m *measurer) gimmeTargets(
	ctx context.Context, sess model.ExperimentSession,
) (map[string]model.TorTarget, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	clnt, err := m.newOrchestraClient(ctx, sess)
	if err != nil {
		return nil, err
	}
	return m.fetchTorTargets(ctx, clnt)
}

// keytarget contains a key and the related target
type keytarget struct {
	key    string
	target model.TorTarget
}

func (kt keytarget) private() bool {
	// TODO(bassosimone): make this dependent on target's value
	return false
}

func (kt keytarget) maybeTargetAddress() string {
	// TODO(bassosimone): make this dependent on target's value
	return kt.target.Address
}

func (m *measurer) measureTargets(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
	targets map[string]model.TorTarget,
) {
	// run measurements in parallel
	var waitgroup sync.WaitGroup
	rc := newResultsCollector(sess, measurement, callbacks)
	waitgroup.Add(len(targets))
	workch := make(chan keytarget)
	for i := 0; i < parallelism; i++ {
		go func(ch <-chan keytarget, total int) {
			for kt := range ch {
				rc.measureSingleTarget(ctx, kt, total)
				waitgroup.Done()
			}
		}(workch, len(targets))
	}
	for key, target := range targets {
		workch <- keytarget{key: key, target: target}
	}
	close(workch)
	waitgroup.Wait()
	// fill the measurement entry
	testkeys := &TestKeys{Targets: rc.targetresults}
	testkeys.fillToplevelKeys()
	measurement.TestKeys = testkeys
}

type resultsCollector struct {
	callbacks       model.ExperimentCallbacks
	completed       *atomicx.Int64
	flexibleConnect func(context.Context, keytarget) (oonitemplates.Results, error)
	measurement     *model.Measurement
	mu              sync.Mutex
	sess            model.ExperimentSession
	targetresults   map[string]TargetResults
}

func newResultsCollector(
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) *resultsCollector {
	rc := &resultsCollector{
		callbacks:     callbacks,
		completed:     atomicx.NewInt64(),
		measurement:   measurement,
		sess:          sess,
		targetresults: make(map[string]TargetResults),
	}
	rc.flexibleConnect = rc.defaultFlexibleConnect
	return rc
}

func maybeSanitize(tr TargetResults, scrub bool) TargetResults {
	// TODO(bassosimone): make this dependent on scrub param
	return tr
}

func (rc *resultsCollector) measureSingleTarget(
	ctx context.Context, kt keytarget, total int,
) {
	tk, err := rc.flexibleConnect(ctx, kt)
	tr := TargetResults{
		Agent:         "redirect",
		Failure:       setFailure(err),
		NetworkEvents: oonidatamodel.NewNetworkEventsList(tk),
		Queries:       oonidatamodel.NewDNSQueriesList(tk),
		Requests:      oonidatamodel.NewRequestList(tk),
		TCPConnect:    oonidatamodel.NewTCPConnectList(tk),
		TLSHandshakes: oonidatamodel.NewTLSHandshakesList(tk),
	}
	tr.fillSummary()
	tr = maybeSanitize(tr, kt.private())
	rc.mu.Lock()
	tr.TargetAddress = kt.maybeTargetAddress()
	tr.TargetName = kt.target.Name
	tr.TargetProtocol = kt.target.Protocol
	rc.targetresults[kt.key] = tr
	rc.mu.Unlock()
	sofar := rc.completed.Add(1)
	percentage := 0.0
	if total > 0 {
		percentage = float64(sofar) / float64(total)
	}
	rc.callbacks.OnProgress(percentage, fmt.Sprintf(
		"tor: access %s/%s: %s", kt.maybeTargetAddress(), kt.target.Protocol,
		errString(err),
	))
}

func maybeScrubbingLogger(logger model.Logger, scrub bool) model.Logger {
	// TODO(bassosimone): make this dependent on scrub
	return logger
}

func (rc *resultsCollector) defaultFlexibleConnect(
	ctx context.Context, kt keytarget,
) (tk oonitemplates.Results, err error) {
	logger := maybeScrubbingLogger(rc.sess.Logger(), kt.private())
	switch kt.target.Protocol {
	case "dir_port":
		url := url.URL{
			Host:   kt.target.Address,
			Path:   "/tor/status-vote/current/consensus.z",
			Scheme: "http",
		}
		const snapshotsize = 1 << 8 // no need to include all in report
		r := oonitemplates.HTTPDo(ctx, oonitemplates.HTTPDoConfig{
			Accept:                  httpheader.RandomAccept(),
			AcceptLanguage:          httpheader.RandomAcceptLanguage(),
			Beginning:               rc.measurement.MeasurementStartTimeSaved,
			MaxEventsBodySnapSize:   snapshotsize,
			MaxResponseBodySnapSize: snapshotsize,
			Handler:                 netxlogger.NewHandler(logger),
			Method:                  "GET",
			URL:                     url.String(),
			UserAgent:               httpheader.RandomUserAgent(),
		})
		tk, err = r.TestKeys, r.Error
	case "or_port", "or_port_dirauth":
		r := oonitemplates.TLSConnect(ctx, oonitemplates.TLSConnectConfig{
			Address:            kt.target.Address,
			Beginning:          rc.measurement.MeasurementStartTimeSaved,
			InsecureSkipVerify: true,
			Handler:            netxlogger.NewHandler(logger),
		})
		tk, err = r.TestKeys, r.Error
	case "obfs4":
		r := oonitemplates.OBFS4Connect(ctx, oonitemplates.OBFS4ConnectConfig{
			Address:      kt.target.Address,
			Beginning:    rc.measurement.MeasurementStartTimeSaved,
			Handler:      netxlogger.NewHandler(logger),
			Params:       kt.target.Params,
			StateBaseDir: rc.sess.TempDir(),
		})
		tk, err = r.TestKeys, r.Error
	default:
		r := oonitemplates.TCPConnect(ctx, oonitemplates.TCPConnectConfig{
			Address:   kt.target.Address,
			Beginning: rc.measurement.MeasurementStartTimeSaved,
			Handler:   netxlogger.NewHandler(logger),
		})
		tk, err = r.TestKeys, r.Error
	}
	return
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return newMeasurer(config)
}

func errString(err error) (s string) {
	s = "success"
	if err != nil {
		s = err.Error()
	}
	return
}

func setFailure(err error) (s *string) {
	if err != nil {
		descr := err.Error()
		s = &descr
	}
	return
}
