// +build !cgo

// Package telegram contains the Telegram network experiment. This file
// in particular is a pure-Go implementation of that.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-020-telegram.md.
package telegram

import (
	"context"
	"net"
	"time"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/urlmeasurer"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "telegram"
	testVersion = "0.0.2"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains telegram test keys.
type TestKeys struct {
	TelegramHTTPBlocking bool                           `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool                           `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string                        `json:"telegram_web_failure"`
	TelegramWebStatus    *string                        `json:"telegram_web_status"`
	XURLMeasurerOutput   map[string]*urlmeasurer.Output `json:"x-url-measurer-output"`
}

func collectMeasurements(
	ctx context.Context, sess *session.Session, testkeys *TestKeys,
) {
	var addresses = []string{
		"149.154.175.50", "149.154.167.51", "149.154.175.100",
		"149.154.167.91", "149.154.171.5",
	}
	measurer := &urlmeasurer.URLMeasurer{
		CABundlePath: sess.CABundlePath(),
		DNSNetwork:   "system",
		Logger:       sess.Logger,
	}
	for _, addr := range addresses {
		URL := "http://" + net.JoinHostPort(addr, "80")
		output := measurer.Do(ctx, urlmeasurer.Input{
			Method: "POST",
			URL:    URL,
		})
		testkeys.XURLMeasurerOutput[URL] = output
		// Note: it's intended to connect using `http` on port `443`. I was
		// surprised as well, but this is the spec and using `https` is actually
		// going to lead to I/O timeouts and other failures.
		URL = "http://" + net.JoinHostPort(addr, "443")
		output = measurer.Do(ctx, urlmeasurer.Input{
			Method: "POST",
			URL:    URL,
		})
	}
	URL := "http://web.telegram.org/"
	output := measurer.Do(ctx, urlmeasurer.Input{
		Method: "GET",
		URL:    URL,
	})
	testkeys.XURLMeasurerOutput[URL] = output
	URL = "https://web.telegram.org/"
	output = measurer.Do(ctx, urlmeasurer.Input{
		Method: "GET",
		URL:    URL,
	})
	testkeys.XURLMeasurerOutput[URL] = output
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	testkeys := &TestKeys{
		TelegramHTTPBlocking: true,
		TelegramTCPBlocking:  true,
		XURLMeasurerOutput:   make(map[string]*urlmeasurer.Output),
	}
	measurement.TestKeys = testkeys
	collectMeasurements(ctx, sess, testkeys)
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
