// Package sniblocking contains the SNI blocking network experiment. This file
// in particular is a pure-Go implementation of that.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-024-sni-blocking.md.
package sniblocking

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	testName    = "sni_blocking"
	testVersion = "0.0.4"
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
	BytesReceived int64                           `json:"-"`
	BytesSent     int64                           `json:"-"`
	Cached        bool                            `json:"-"`
	Failure       *string                         `json:"failure"`
	NetworkEvents oonidatamodel.NetworkEventsList `json:"network_events"`
	Queries       oonidatamodel.DNSQueriesList    `json:"queries"`
	Requests      oonidatamodel.RequestList       `json:"requests"`
	SNI           string                          `json:"sni"`
	TCPConnect    oonidatamodel.TCPConnectList    `json:"tcp_connect"`
	THAddress     string                          `json:"th_address"`
	TLSHandshakes oonidatamodel.TLSHandshakesList `json:"tls_handshakes"`
}

// TestKeys contains sniblocking test keys.
type TestKeys struct {
	Result  string    `json:"result"`
	Control Subresult `json:"control"`
	Target  Subresult `json:"target"`
}

const (
	classAccessibleInvalidHostname = "accessible_invalid_hostname"
	classAccessibleValidHostname   = "accessible_valid_hostname"
	classAnomalySSLError           = "anomaly_ssl_error"
	classAnomalyTestHelperBlocked  = "anomaly_test_helper_blocked"
	classAnomalyTimeout            = "anomaly_timeout"
	classAnomalyUnexpectedFailure  = "anomaly_unexpected_failure"
	classBlockedTCPIPError         = "blocked_tcpip_error"
)

func (tk *TestKeys) classify() string {
	if tk.Target.Failure == nil {
		return classAccessibleValidHostname
	}
	switch *tk.Target.Failure {
	case modelx.FailureConnectionRefused:
		return classAnomalyTestHelperBlocked
	case modelx.FailureDNSNXDOMAINError:
		return classAnomalyTestHelperBlocked
	case modelx.FailureConnectionReset:
		return classBlockedTCPIPError
	case modelx.FailureEOFError:
		return classBlockedTCPIPError
	case modelx.FailureSSLInvalidHostname:
		return classAccessibleInvalidHostname
	case modelx.FailureSSLUnknownAuthority:
		return classAnomalySSLError
	case modelx.FailureSSLInvalidCertificate:
		return classAnomalySSLError
	case modelx.FailureGenericTimeoutError:
		if tk.Control.Failure != nil {
			return classAnomalyTestHelperBlocked
		}
		return classAnomalyTimeout
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
	handler modelx.Handler,
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
		s := modelx.FailureGenericTimeoutError
		return Subresult{
			Failure: &s,
			SNI:     sni,
		}
	}
	// perform the measurement
	result := oonitemplates.TLSConnect(ctx, oonitemplates.TLSConnectConfig{
		Address:   thaddr,
		Beginning: beginning,
		Handler:   handler,
		SNI:       sni,
	})
	// assemble and publish the results
	smk := Subresult{
		BytesReceived: result.TestKeys.ReceivedBytes,
		BytesSent:     result.TestKeys.SentBytes,
		NetworkEvents: oonidatamodel.NewNetworkEventsList(result.TestKeys),
		Queries:       oonidatamodel.NewDNSQueriesList(result.TestKeys),
		Requests:      oonidatamodel.NewRequestList(result.TestKeys),
		SNI:           sni,
		TCPConnect:    oonidatamodel.NewTCPConnectList(result.TestKeys),
		THAddress:     thaddr,
		TLSHandshakes: oonidatamodel.NewTLSHandshakesList(result.TestKeys),
	}
	if result.Error != nil {
		s := result.Error.Error()
		smk.Failure = &s
	}
	return smk
}

func (m *measurer) measureonewithcache(
	ctx context.Context,
	output chan<- Subresult,
	handler modelx.Handler,
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
	smk = m.measureone(ctx, handler, beginning, sni, thaddr)
	output <- smk
	smk.BytesReceived = 0 // don't count them more than once
	smk.BytesSent = 0     // ditto
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
			ctx, outputs, netxlogger.NewHandler(sess.Logger()),
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
		current       int
		sentBytes     int64
		receivedBytes int64
		testkeys      = new(TestKeys)
	)
	for smk := range outputs {
		if smk.SNI == controlSNI {
			testkeys.Control = smk
		} else if smk.SNI == measurement.Input {
			testkeys.Target = smk
		} else {
			panic("unexpected smk.SNI")
		}
		sentBytes += smk.BytesSent
		receivedBytes += smk.BytesReceived
		current++
		sess.Logger().Infof(
			"sni_blocking: %s: %s [cached: %+v]", smk.SNI,
			asString(smk.Failure), smk.Cached)
		if current >= len(inputs) {
			break
		}
	}
	testkeys.Result = testkeys.classify()
	sess.Logger().Infof("sni_blocking: result: %s", testkeys.Result)
	callbacks.OnDataUsage(
		float64(receivedBytes)/1024.0, // downloaded
		float64(sentBytes)/1024.0,     // uploaded
	)
	return testkeys
}

// maybeURLToSNI handles the case where the input is from the test-lists
// and hence every input is a URL rather than a domain.
func maybeURLToSNI(input string) (string, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return "", err
	}
	if parsed.Path == input {
		return input, nil
	}
	return parsed.Hostname(), nil
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
	maybeParsed, err := maybeURLToSNI(measurement.Input)
	if err != nil {
		return err
	}
	measurement.Input = maybeParsed
	inputs := []string{m.config.ControlSNI}
	if measurement.Input != m.config.ControlSNI {
		inputs = append(inputs, measurement.Input)
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
