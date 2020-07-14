// Package webconnectivity implements OONI's Web Connectivity experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-017-web-connectivity.md
package webconnectivity

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/archival"
)

const (
	testName    = "web_connectivity"
	testVersion = "0.1.0"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains webconnectivity test keys.
type TestKeys struct {
	// measurement from urlgetter
	urlgetter.TestKeys

	// other fields
	ClientResolver        string  `json:"client_resolver"`
	Retries               *int64  `json:"retries"` // unused
	DNSExperimentFailure  *string `json:"dns_experiment_failure"`
	HTTPExperimentFailure *string `json:"http_experiment_failure"`

	// control
	ControlFailure *string         `json:"control_failure"`
	ControlRequest ControlRequest  `json:"-"`
	Control        ControlResponse `json:"control"`

	// analysis
	AnalysisResult
}

// Measurer performs the measurement.
type Measurer struct {
	Config Config
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return Measurer{Config: config}
}

// ExperimentName implements ExperimentMeasurer.ExperExperimentName.
func (m Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion implements ExperimentMeasurer.ExperExperimentVersion.
func (m Measurer) ExperimentVersion() string {
	return testVersion
}

var (
	// ErrNoAvailableTestHelpers is emitted when there are no available test helpers.
	ErrNoAvailableTestHelpers = errors.New("no available helpers")

	// ErrNoInput indicates that no input was provided
	ErrNoInput = errors.New("no input provided")

	// ErrUnsupportedInput indicates that the input URL scheme is unsupported.
	ErrUnsupportedInput = errors.New("unsupported input scheme")
)

// Run implements ExperimentMeasurer.Run.
func (m Measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	urlgetter.RegisterExtensions(measurement)
	tk := new(TestKeys)
	measurement.TestKeys = tk
	tk.ClientResolver = sess.ResolverIP()
	if measurement.Input == "" {
		return ErrNoInput
	}
	if strings.HasPrefix(string(measurement.Input), "http://") != true &&
		strings.HasPrefix(string(measurement.Input), "https://") != true {
		return ErrUnsupportedInput
	}
	// 1. find test helper
	testhelpers, _ := sess.GetTestHelpersByName("web-connectivity")
	var testhelper *model.Service
	for _, th := range testhelpers {
		if th.Type == "https" {
			testhelper = &th
			break
		}
	}
	if testhelper == nil {
		return ErrNoAvailableTestHelpers
	}
	measurement.TestHelpers = map[string]interface{}{
		"backend": testhelper,
	}
	// 2. perform the measurement
	tk.TestKeys = Measure(ctx, sess, measurement.Input)
	tk.DNSExperimentFailure = DNSExperimentFailure(tk)
	tk.HTTPExperimentFailure = HTTPExperimentFailure(tk)
	// 3. contact the control
	tk.ControlRequest = NewControlRequest(measurement.Input, tk.TestKeys)
	var err error
	tk.Control, err = Control(ctx, sess, testhelper.Address, tk.ControlRequest)
	tk.ControlFailure = archival.NewFailure(err)
	// 4. rewrite TCPConnect to include blocking information - it is very
	// sad that we're storing analysis result inside the measurement
	tk.TCPConnect = ComputeTCPBlocking(tk.TCPConnect, tk.Control.TCPConnect)
	// 5. compare measurement to control
	tk.AnalysisResult = Analyze(string(measurement.Input), tk)
	return nil
}

// ComputeTCPBlocking will return a copy of the input TCPConnect structure
// where we set the Blocking value depending on the control results.
func ComputeTCPBlocking(measurement []archival.TCPConnectEntry,
	control map[string]ControlTCPConnectResult) (out []archival.TCPConnectEntry) {
	out = []archival.TCPConnectEntry{}
	for _, me := range measurement {
		epnt := net.JoinHostPort(me.IP, strconv.Itoa(me.Port))
		if ce, ok := control[epnt]; ok {
			v := ce.Failure == nil && me.Status.Failure != nil
			me.Status.Blocked = &v
		}
		out = append(out, me)
	}
	return
}
