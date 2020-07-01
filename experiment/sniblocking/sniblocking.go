// Package sniblocking contains the SNI blocking network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-024-sni-blocking.md.
package sniblocking

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	testName    = "sni_blocking"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct {
	// ControlSNI is the SNI to be used for the control.
	ControlSNI string

	// TestHelperAddress is the address of the test helper.
	TestHelperAddress string
}

// Subresult contains the keys of a single measurement
// that targets either the target or the control.
type Subresult struct {
	urlgetter.TestKeys
	Cached    bool   `json:"-"`
	SNI       string `json:"sni"`
	THAddress string `json:"th_address"`
}

// TestKeys contains sniblocking test keys.
type TestKeys struct {
	Control Subresult `json:"control"`
	Result  string    `json:"result"`
	Target  Subresult `json:"target"`
}

const (
	classAnomalyTestHelperUnreachable   = "anomaly.test_helper_unreachable"
	classAnomalyTimeout                 = "anomaly.timeout"
	classAnomalyUnexpectedFailure       = "anomaly.unexpected_failure"
	classInterferenceClosed             = "interference.closed"
	classInterferenceInvalidCertificate = "interference.invalid_certificate"
	classInterferenceReset              = "interference.reset"
	classInterferenceUnknownAuthority   = "interference.unknown_authority"
	classSuccessGotServerHello          = "success.got_server_hello"
)

func (tk *TestKeys) classify() string {
	if tk.Target.Failure == nil {
		return classSuccessGotServerHello
	}
	switch *tk.Target.Failure {
	case modelx.FailureConnectionRefused:
		return classAnomalyTestHelperUnreachable
	case modelx.FailureConnectionReset:
		return classInterferenceReset
	case modelx.FailureDNSNXDOMAINError:
		return classAnomalyTestHelperUnreachable
	case modelx.FailureEOFError:
		return classInterferenceClosed
	case modelx.FailureGenericTimeoutError:
		if tk.Control.Failure != nil {
			return classAnomalyTestHelperUnreachable
		}
		return classAnomalyTimeout
	case modelx.FailureSSLInvalidCertificate:
		return classInterferenceInvalidCertificate
	case modelx.FailureSSLInvalidHostname:
		return classSuccessGotServerHello
	case modelx.FailureSSLUnknownAuthority:
		return classInterferenceUnknownAuthority
	}
	return classAnomalyUnexpectedFailure
}

type measurer struct {
	cache  map[string]Subresult
	config Config
	mu     sync.Mutex
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) measureone(
	ctx context.Context,
	sess model.ExperimentSession,
	beginning time.Time,
	sni string,
	thaddr string,
) Subresult {
	// slightly delay the measurement
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	sleeptime := time.Duration(gen.Intn(250)) * time.Millisecond
	select {
	case <-time.After(sleeptime):
	case <-ctx.Done():
		s := modelx.FailureInterrupted
		return Subresult{
			TestKeys: urlgetter.TestKeys{
				Failure: &s,
			},
			THAddress: thaddr,
			SNI:       sni,
		}
	}
	// perform the measurement
	g := urlgetter.Getter{
		Begin:   beginning,
		Config:  urlgetter.Config{TLSServerName: sni},
		Session: sess,
		Target:  fmt.Sprintf("tlshandshake://%s", thaddr),
	}
	// Ignoring the error because g.Get() sets the tk.Failure field
	// to be the OONI equivalent of the error that occurred.
	tk, _ := g.Get(ctx)
	// assemble and publish the results
	smk := Subresult{
		SNI:       sni,
		THAddress: thaddr,
		TestKeys:  tk,
	}
	return smk
}

func (m *measurer) measureonewithcache(
	ctx context.Context,
	output chan<- Subresult,
	sess model.ExperimentSession,
	beginning time.Time,
	sni string,
	thaddr string,
) {
	cachekey := sni + thaddr
	m.mu.Lock()
	smk, okay := m.cache[cachekey]
	m.mu.Unlock()
	if okay {
		output <- smk
		return
	}
	smk = m.measureone(ctx, sess, beginning, sni, thaddr)
	output <- smk
	smk.Cached = true
	m.mu.Lock()
	m.cache[cachekey] = smk
	m.mu.Unlock()
}

func (m *measurer) startall(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, inputs []string,
) <-chan Subresult {
	outputs := make(chan Subresult, len(inputs))
	for _, input := range inputs {
		go m.measureonewithcache(
			ctx, outputs, sess,
			measurement.MeasurementStartTimeSaved,
			input, m.config.TestHelperAddress,
		)
	}
	return outputs
}

func processall(
	outputs <-chan Subresult,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
	inputs []string,
	sess model.ExperimentSession,
	controlSNI string,
) *TestKeys {
	var (
		current  int
		testkeys = new(TestKeys)
	)
	for smk := range outputs {
		if smk.SNI == controlSNI {
			testkeys.Control = smk
		} else if smk.SNI == string(measurement.Input) {
			testkeys.Target = smk
		} else {
			panic("unexpected smk.SNI")
		}
		current++
		sess.Logger().Debugf(
			"sni_blocking: %s: %s [cached: %+v]", smk.SNI,
			asString(smk.Failure), smk.Cached)
		if current >= len(inputs) {
			break
		}
	}
	testkeys.Result = testkeys.classify()
	sess.Logger().Infof("sni_blocking: result: %s", testkeys.Result)
	return testkeys
}

// maybeURLToSNI handles the case where the input is from the test-lists
// and hence every input is a URL rather than a domain.
func maybeURLToSNI(input model.MeasurementTarget) (model.MeasurementTarget, error) {
	parsed, err := url.Parse(string(input))
	if err != nil {
		return "", err
	}
	if parsed.Path == string(input) {
		return input, nil
	}
	return model.MeasurementTarget(parsed.Hostname()), nil
}

func (m *measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) error {
	m.mu.Lock()
	if m.cache == nil {
		m.cache = make(map[string]Subresult)
	}
	m.mu.Unlock()
	if m.config.ControlSNI == "" {
		return errors.New("Experiment requires ControlSNI")
	}
	if measurement.Input == "" {
		return errors.New("Experiment requires measurement.Input")
	}
	if m.config.TestHelperAddress == "" {
		m.config.TestHelperAddress = net.JoinHostPort(
			m.config.ControlSNI, "443",
		)
	}
	urlgetter.RegisterExtensions(measurement)
	// TODO(bassosimone): if the user has configured DoT or DoH, here we
	// probably want to perform the name resolution before the measurements
	// or to make sure that the classify logic is robust to that.
	//
	// See https://github.com/ooni/probe-engine/issues/392.
	maybeParsed, err := maybeURLToSNI(measurement.Input)
	if err != nil {
		return err
	}
	measurement.Input = maybeParsed
	inputs := []string{m.config.ControlSNI}
	if string(measurement.Input) != m.config.ControlSNI {
		inputs = append(inputs, string(measurement.Input))
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second*time.Duration(len(inputs)))
	defer cancel()
	outputs := m.startall(ctx, sess, measurement, inputs)
	measurement.TestKeys = processall(
		outputs, measurement, callbacks, inputs, sess, m.config.ControlSNI,
	)
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}

func asString(failure *string) (result string) {
	result = "success"
	if failure != nil {
		result = *failure
	}
	return
}
