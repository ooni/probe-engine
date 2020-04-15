// Package dash contains the dash network experiment. This file
// in particular is a pure-Go implementation of this test.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-021-dash.md
package dash

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
)

const (
	testName       = "dash"
	testVersion    = "0.8.0"
	defaultTimeout = 55 * time.Second
	totalStep      = 15.0
)

// Config contains the experiment config.
type Config struct{}

// Simple contains the experiment total summary
type Simple struct {
	ConnectLatency  float64 `json:"connect_latency"`
	MedianBitrate   int64   `json:"median_bitrate"`
	MinPlayoutDelay float64 `json:"min_playout_delay"`
}

// TestKeys contains the test keys
type TestKeys struct {
	Simple       Simple          `json:"simple"`
	Failure      *string         `json:"failure"`
	ReceiverData []clientResults `json:"receiver_data"`
}

type dashClient interface {
	StartDownload(ctx context.Context) (<-chan clientResults, error)
	Error() error
	ServerResults() []serverResults
}

type runner struct {
	callbacks   model.ExperimentCallbacks
	clnt        dashClient
	jsonMarshal func(v interface{}) ([]byte, error)
	logger      model.Logger
	tk          *TestKeys
}

func newRunner(
	logger model.Logger, clnt dashClient,
	callbacks model.ExperimentCallbacks,
	jsonMarshal func(v interface{}) ([]byte, error),
) *runner {
	return &runner{
		callbacks:   callbacks,
		clnt:        clnt,
		jsonMarshal: jsonMarshal,
		logger:      logger,
		tk:          new(TestKeys),
	}
}

// loop runs the neubot/dash measurement loop and writes
// interim results of the test in `tk`. It is not this
// function's concern to set tk.Failure. The caller must do it
// when this function returns a non-nil error.
func (r *runner) loop(ctx context.Context) error {
	ch, err := r.clnt.StartDownload(ctx)
	if err != nil {
		return err
	}
	percentage := 0.0
	step := 1 / (totalStep + 1)
	for results := range ch {
		percentage += step
		message := fmt.Sprintf(
			"rate: %s/s speed: %s/s elapsed: %.2f s",
			humanize.Bytes(uint64(results.Rate*1000/8)), // Rate is kbit in SI size
			humanize.Bytes(uint64(float64(results.Received)/results.Elapsed)),
			results.Elapsed,
		)
		r.callbacks.OnProgress(percentage, message)
		data, err := r.jsonMarshal(results)
		if err != nil {
			return err
		}
		r.logger.Debugf("%s", string(data))
		r.tk.ReceiverData = append(r.tk.ReceiverData, results)
	}
	if r.clnt.Error() != nil {
		return r.clnt.Error()
	}
	data, err := r.jsonMarshal(r.clnt.ServerResults())
	if err != nil {
		return err
	}
	r.logger.Debugf("Server result: %s", string(data))
	// TODO(bassosimone): it seems we're not saving the server data?
	return nil
}

// analyze analyzes the results of DASH and fills stats inside of tk.
func (tk *TestKeys) analyze() error {
	var rates []float64
	var frameReadyTime float64
	var playTime float64
	for _, results := range tk.ReceiverData {
		rates = append(rates, float64(results.Rate))
		tk.Simple.ConnectLatency = results.ConnectTime // same in all samples
		// Rationale: first segment plays when it arrives. Subsequent segments
		// would play in ElapsedTarget seconds. However, will play when they
		// arrive. Stall is the time we need to wait for a frame to arrive with
		// the video stopped and the spinning icon.
		frameReadyTime += results.Elapsed
		if playTime == 0.0 {
			playTime += frameReadyTime
		} else {
			playTime += float64(results.ElapsedTarget)
		}
		stall := frameReadyTime - playTime
		if stall > tk.Simple.MinPlayoutDelay {
			tk.Simple.MinPlayoutDelay = stall
		}
	}
	median, err := stats.Median(rates)
	tk.Simple.MedianBitrate = int64(median)
	return err
}

// printSummary just prints a debug-level summary. We cannot use the info
// level because that is reserved for the OONI Probe CLI.
func (tk *TestKeys) printSummary(logger model.Logger) {
	logger.Debugf("Test Summary: ")
	logger.Debugf("Connect latency: %s",
		// convert to nanoseconds
		time.Duration(tk.Simple.ConnectLatency*1000000000),
	)
	logger.Debugf("Median bitrate: %s/s",
		// MedianBitrate is kbit in SI size
		humanize.Bytes(uint64(tk.Simple.MedianBitrate*1000/8)),
	)
	logger.Debugf("Min. playout delay: %.3f s", tk.Simple.MinPlayoutDelay)
}

// do is the main function of the runner
func (r *runner) do(ctx context.Context) error {
	err := r.loop(ctx)
	if err != nil {
		s := err.Error()
		r.tk.Failure = &s
		return err
	}
	err = r.tk.analyze()
	if err != nil {
		s := err.Error()
		r.tk.Failure = &s
		return err
	}
	r.callbacks.OnProgress(1, "done")
	r.tk.printSummary(r.logger)
	return nil
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

func (m *measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	dlr := dialer.DNSDialer{
		Dialer: dialer.LoggingDialer{
			Dialer: dialer.ErrorWrapperDialer{
				Dialer: dialer.TimeoutDialer{
					Dialer: dialer.ByteCounterDialer{
						Dialer: new(net.Dialer),
					},
				},
			},
			Logger: sess.Logger(),
		},
		Resolver: new(net.Resolver),
	}
	tlsdlr := dialer.TLSDialer{
		Dialer: dlr,
		TLSHandshaker: dialer.LoggingTLSHandshaker{
			TLSHandshaker: dialer.ErrorWrapperTLSHandshaker{
				TLSHandshaker: dialer.TimeoutTLSHandshaker{
					TLSHandshaker: dialer.SystemTLSHandshaker{},
				},
			},
			Logger: sess.Logger(),
		},
	}
	httpClient := &http.Client{
		Transport: httptransport.LoggingTransport{
			RoundTripper: httptransport.UserAgentTransport{
				RoundTripper: httptransport.NewSystemTransport(
					dlr, tlsdlr, nil,
				),
			},
			Logger: sess.Logger(),
		},
	}
	defer httpClient.CloseIdleConnections()
	clnt := newClient(httpClient, sess.Logger(), callbacks,
		sess.SoftwareName(), sess.SoftwareVersion())
	r := newRunner(sess.Logger(), clnt, callbacks, json.Marshal)
	measurement.TestKeys = r.tk
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	return r.do(ctx)
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}
