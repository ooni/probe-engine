// Package dash contains the dash network experiment. This file
// in particular is a pure-Go implementation of this test.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-021-dash.md
package dash

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/ooni/probe-engine/internal/humanizex"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	defaultTimeout = 120 * time.Second
	testName       = "dash"
	totalStep      = 15.0
	version        = 11
)

var (
	errServerBusy        = errors.New("Server busy; try again later")
	errHTTPRequestFailed = errors.New("HTTP request failed")
	magicVersion         = fmt.Sprintf("0.%03d000000", version)
	testVersion          = fmt.Sprintf("0.%d.0", version)
)

// Config contains the experiment config.
type Config struct {
	Tunnel string `ooni:"Run experiment over a tunnel, e.g. psiphon"`
}

// Simple contains the experiment total summary
type Simple struct {
	ConnectLatency  float64 `json:"connect_latency"`
	MedianBitrate   int64   `json:"median_bitrate"`
	MinPlayoutDelay float64 `json:"min_playout_delay"`
}

// TestKeys contains the test keys
type TestKeys struct {
	BootstrapTime float64         `json:"bootstrap_time,omitempty"`
	Simple        Simple          `json:"simple"`
	Failure       *string         `json:"failure"`
	ReceiverData  []clientResults `json:"receiver_data"`
	SOCKSProxy    string          `json:"socksproxy,omitempty"`
	Tunnel        string          `json:"tunnel,omitempty"`
}

type runner struct {
	callbacks  model.ExperimentCallbacks
	httpClient *http.Client
	saver      *trace.Saver
	sess       model.ExperimentSession
	tk         *TestKeys
}

func (r runner) HTTPClient() *http.Client {
	return r.httpClient
}

func (r runner) JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (r runner) Logger() model.Logger {
	return r.sess.Logger()
}

func (r runner) NewHTTPRequest(meth, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(meth, url, body)
}

func (r runner) ReadAll(reader io.Reader) ([]byte, error) {
	return ioutil.ReadAll(reader)
}

func (r runner) Scheme() string {
	return "https"
}

func (r runner) UserAgent() string {
	return r.sess.UserAgent()
}

func (r runner) loop(ctx context.Context, numIterations int64) error {
	fqdn, err := locate(ctx, r)
	if err != nil {
		return err
	}
	r.callbacks.OnProgress(0.0, fmt.Sprintf("streaming: server: %s", fqdn))
	negotiateResp, err := negotiate(ctx, fqdn, r)
	if err != nil {
		return err
	}
	if err := r.measure(ctx, fqdn, negotiateResp, numIterations); err != nil {
		return err
	}
	// TODO(bassosimone): it seems we're not saving the server data?
	err = collect(ctx, fqdn, negotiateResp.Authorization, r.tk.ReceiverData, r)
	if err != nil {
		return err
	}
	return r.tk.analyze()
}

func (r runner) measure(
	ctx context.Context, fqdn string, negotiateResp negotiateResponse,
	numIterations int64) error {
	// Note: according to a comment in MK sources 3000 kbit/s was the
	// minimum speed recommended by Netflix for SD quality in 2017.
	//
	// See: <https://help.netflix.com/en/node/306>.
	const initialBitrate = 3000
	current := clientResults{
		ElapsedTarget: 2,
		Platform:      runtime.GOOS,
		Rate:          initialBitrate,
		RealAddress:   negotiateResp.RealAddress,
		Version:       magicVersion,
	}
	var (
		begin       = time.Now()
		connectTime float64
		ch          = make(chan clientResults)
	)
	go r.player(ctx, ch, numIterations)
	for current.Iteration < numIterations {
		result, err := download(ctx, downloadConfig{
			authorization: negotiateResp.Authorization,
			begin:         begin,
			currentRate:   current.Rate,
			deps:          r,
			elapsedTarget: current.ElapsedTarget,
			fqdn:          fqdn,
		})
		if err != nil {
			// Implementation note: ndt7 controls the connection much
			// more than us and it can tell whether an error occurs when
			// connecting or later. We cannot say that very precisely
			// because, in principle, we may reconnect. So we always
			// return error here. This comment is being introduced so
			// that we don't do https://github.com/ooni/probe-engine/pull/526
			// again, because that isn't accurate.
			return err
		}
		current.Elapsed = result.elapsed
		current.Received = result.received
		current.RequestTicks = result.requestTicks
		current.Timestamp = result.timestamp
		current.ServerURL = result.serverURL
		// Read the events so far and possibly update our measurement
		// of the latest connect time. We should have one sample in most
		// cases, because the connection should be persistent.
		for _, ev := range r.saver.Read() {
			if ev.Name == "connect" {
				connectTime = ev.Duration.Seconds()
			}
		}
		current.ConnectTime = connectTime
		r.tk.ReceiverData = append(r.tk.ReceiverData, current)
		ch <- current
		current.Iteration++
		speed := float64(current.Received) / float64(current.Elapsed)
		speed *= 8.0    // to bits per second
		speed /= 1000.0 // to kbit/s
		current.Rate = int64(speed)
	}
	return nil
}

func (r runner) player(ctx context.Context, frames <-chan clientResults, numIterations int64) {
	var frame clientResults
	// receive the first frame
	select {
	case <-ctx.Done():
		return
	case frame = <-frames:
	}
	for {
		// play the current frame - note that the play rate is the speed at which
		// we extract bytes from the queue, while current.Rate is the rate at which
		// at which we are downloading and filling the queue.
		percentage := float64(frame.Iteration) / float64(numIterations)
		rate := 8 * float64(frame.Received) / float64(frame.ElapsedTarget)
		message := fmt.Sprintf("streaming: play at rate %s", humanizex.SI(rate, "bit/s"))
		r.callbacks.OnProgress(percentage, message)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(frame.ElapsedTarget) * time.Second):
		}
		// receive the next frame
		select {
		case <-ctx.Done():
			return
		case frame = <-frames:
		default:
			r.Logger().Info("streaming: player is stalled")
			select {
			case <-ctx.Done():
				return
			case frame = <-frames:
			}
		}
	}
}

func (tk *TestKeys) analyze() error {
	var (
		rates          []float64
		frameReadyTime float64
		playTime       float64
	)
	for _, results := range tk.ReceiverData {
		rates = append(rates, float64(results.Rate))
		// Same in all samples if we're using a single connection
		tk.Simple.ConnectLatency = results.ConnectTime
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

func (r runner) do(ctx context.Context) error {
	defer r.callbacks.OnProgress(1, "streaming: done")
	const numIterations = 15
	err := r.loop(ctx, numIterations)
	if err != nil {
		s := err.Error()
		r.tk.Failure = &s
		// fallthrough
	}
	return err
}

type measurer struct {
	config Config
}

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	tk := new(TestKeys)
	measurement.TestKeys = tk
	tk.Tunnel = m.config.Tunnel
	if err := sess.MaybeStartTunnel(ctx, m.config.Tunnel); err != nil {
		s := err.Error()
		tk.Failure = &s
		return err
	}
	tk.BootstrapTime = sess.TunnelBootstrapTime().Seconds()
	if url := sess.ProxyURL(); url != nil {
		tk.SOCKSProxy = url.Host
	}
	saver := &trace.Saver{}
	httpClient := &http.Client{
		Transport: httptransport.New(httptransport.Config{
			ContextByteCounting: true,
			Logger:              sess.Logger(),
			ProxyURL:            sess.ProxyURL(),
			Saver:               saver,
		}),
	}
	defer httpClient.CloseIdleConnections()
	r := runner{
		callbacks:  callbacks,
		httpClient: httpClient,
		saver:      saver,
		sess:       sess,
		tk:         tk,
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	return r.do(ctx)
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
