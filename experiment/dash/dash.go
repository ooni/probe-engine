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
	"time"

	"github.com/montanaflynn/stats"
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
	ConnectLatency      float64 `json:"connect_latency"`
	MedianBitratePlayer int64   `json:"median_bitrate_player"`
	MedianBitrate       int64   `json:"median_bitrate"`
	MinPlayoutDelay     float64 `json:"min_playout_delay"`
}

// ServerInfo contains information on the selected server
//
// This is currently an extension to the DASH specification
// until the data format of the new mlab locate is clear.
type ServerInfo struct {
	Hostname string `json:"hostname"`
	Site     string `json:"site,omitempty"`
}

// TestKeys contains the test keys
type TestKeys struct {
	BootstrapTime float64         `json:"bootstrap_time,omitempty"`
	Failure       *string         `json:"failure"`
	PlayerData    []playerInfo    `json:"player_data"`
	ReceiverData  []clientResults `json:"receiver_data"`
	SOCKSProxy    string          `json:"socksproxy,omitempty"`
	Simple        Simple          `json:"simple"`
	Server        ServerInfo      `json:"server"`
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
	locateResult, err := locate(ctx, r)
	if err != nil {
		return err
	}
	r.tk.Server = ServerInfo{
		Hostname: locateResult.FQDN,
		Site:     locateResult.Site,
	}
	fqdn := locateResult.FQDN
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
	return r.play(ctx, playConfig{
		authorization: negotiateResp.Authorization,
		fqdn:          fqdn,
		numIterations: numIterations,
		realAddress:   negotiateResp.RealAddress,
	})
}

func (tk *TestKeys) analyze() error {
	var rates []float64
	for _, results := range tk.ReceiverData {
		rates = append(rates, float64(results.Rate))
		// Same in all samples if we're using a single connection
		tk.Simple.ConnectLatency = results.ConnectTime
	}
	median, _ := stats.Median(rates)
	tk.Simple.MedianBitrate = int64(median)
	return tk.analyzePlayer()
}

func (tk *TestKeys) analyzePlayer() error {
	var playRates []float64
	for _, results := range tk.PlayerData {
		playRates = append(playRates, float64(results.Rate))
	}
	median, _ := stats.Median(playRates)
	tk.Simple.MedianBitratePlayer = int64(median)
	return nil
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
