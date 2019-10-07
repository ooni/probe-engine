// +build !cgo

// Package dash contains the dash network experiment. This file
// in particular is a pure-Go implementation of that.
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
	Simple Simple `json:"simple"`

	// Failure is the failure string
	Failure string `json:"failure"`

	ReceiverData []neubotModel.ClientResults `json:"receiver_data"`
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
	// StartDownload starts the DASH download. It returns a channel where
	// client measurements are posted, or an error.
	ch, err := client.StartDownload(ctx)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	callbacks.OnProgress(0, fmt.Sprintf("server: %s", client.FQDN))
	var rates []float64
	var frameReadyTime float64 = 0.0
	var playTime float64 = 0.0
	percentage := 0.0
	step := 100 / (totalStep + 1) / 100
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
			testkeys.Failure = err.Error()
			return err
		}
		sess.Logger.Debugf("%s", string(data))

		// Here we're computing stats inline
		// rates is used to calculate MedianBitrate.
		rates = append(rates, float64(results.Rate))

		if testkeys.Simple.ConnectLatency == 0.0 {
			// It is always equal for all the records
			testkeys.Simple.ConnectLatency = results.ConnectTime
		}

		frameReadyTime += float64(results.Elapsed)
		if playTime == 0 {
			playTime += frameReadyTime
		} else {
			playTime += float64(results.ElapsedTarget)
		}
		var stall float64 = frameReadyTime - playTime
		if stall > testkeys.Simple.MinPlayoutDelay {
			testkeys.Simple.MinPlayoutDelay = stall
		}
		testkeys.ReceiverData = append(testkeys.ReceiverData, results)
	}

	median, err := stats.Median(rates)
	testkeys.Simple.MedianBitrate = int64(median)
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	if client.Error() != nil {
		testkeys.Failure = err.Error()
		return client.Error()
	}
	data, err := json.Marshal(client.ServerResults())
	if err != nil {
		testkeys.Failure = err.Error()
		return err
	}
	sess.Logger.Debugf("Server result: %s", string(data))
	callbacks.OnProgress(1, "done")
	sess.Logger.Debugf("Test Summary: ")
	sess.Logger.Debugf("Connect latency: %s",
		// convert to nanoseconds
		time.Duration(testkeys.Simple.ConnectLatency*1000000000),
	)
	sess.Logger.Debugf("Median bitrate: %s/s",
		// MedianBitrate is kbit in SI size
		humanize.IBytes(uint64(testkeys.Simple.MedianBitrate*1000/8)),
	)
	sess.Logger.Debugf("Min. playout delay: %.3f s", testkeys.Simple.MinPlayoutDelay)
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
