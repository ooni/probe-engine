// +build !cgo

// Package dash contains the dash network experiment. This file
// in particular is a pure-Go implementation of this test.
//
// Spec: https://github.com/ooni/spec/blob/master/nettests/ts-021-dash.md
package dash

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/neubot/dash/client"
	neubotModel "github.com/neubot/dash/model"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
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
	Simple       Simple                      `json:"simple"`
	Failure      string                      `json:"failure"`
	ReceiverData []neubotModel.ClientResults `json:"receiver_data"`
}

// loop runs the neubot/dash measurement loop and writes the
// interim all the results of the test in `tk`. It is not this
// function concern to set tk.Failure. The caller must do it
// when this function returns a non-nil error.
func (tk *TestKeys) loop(
	ctx context.Context, sess *session.Session,
	client *client.Client, callbacks handler.Callbacks,
) error {
	ch, err := client.StartDownload(ctx)
	if err != nil {
		return err
	}
	callbacks.OnProgress(0, fmt.Sprintf("server: %s", client.FQDN))
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
		callbacks.OnProgress(percentage, message)
		data, err := json.Marshal(results)
		if err != nil {
			return err
		}
		sess.Logger.Debugf("%s", string(data))
		tk.ReceiverData = append(tk.ReceiverData, results)
	}
	if client.Error() != nil {
		return err
	}
	data, err := json.Marshal(client.ServerResults())
	if err != nil {
		return err
	}
	sess.Logger.Debugf("Server result: %s", string(data))
	// TODO(bassosimone): it seems we're not saving the server data?
	return nil
}

// analyze analyzes the results of DASH and fills stats inside of tk.
func (tk *TestKeys) analyze(
	sess *session.Session, client *client.Client, callbacks handler.Callbacks,
) error {
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
		frameReadyTime += float64(results.Elapsed)
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
func (tk *TestKeys) printSummary(sess *session.Session) {
	sess.Logger.Debugf("Test Summary: ")
	sess.Logger.Debugf("Connect latency: %s",
		// convert to nanoseconds
		time.Duration(tk.Simple.ConnectLatency*1000000000),
	)
	sess.Logger.Debugf("Median bitrate: %s/s",
		// MedianBitrate is kbit in SI size
		humanize.Bytes(uint64(tk.Simple.MedianBitrate*1000/8)),
	)
	sess.Logger.Debugf("Min. playout delay: %.3f s", tk.Simple.MinPlayoutDelay)
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	testkeys := &TestKeys{}
	measurement.TestKeys = testkeys
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	client := client.New(sess.SoftwareName, sess.SoftwareVersion)
	err := testkeys.loop(ctx, sess, client, callbacks)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	err = testkeys.analyze(sess, client, callbacks)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	callbacks.OnProgress(1, "done")
	testkeys.printSummary(sess)
	return nil
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(
		sess, testName, testVersion,
		func(
			ctx context.Context,
			sess *session.Session,
			measurement *model.Measurement,
			callbacks handler.Callbacks,
		) error {
			return measure(ctx, sess, measurement, callbacks, config)
		})
}
