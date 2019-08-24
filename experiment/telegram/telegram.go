// Package telegram contains the Telegram network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-020-telegram.md.
package telegram

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/ootemplate"
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
	TCPConnect           []ootemplate.TCPConnectResults `json:"tcp_connect"`
	Requests             []ootemplate.HTTPRoundTrip     `json:"requests"`
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
	}
	measurement.TestKeys = testkeys
	defer func() {
		for _, tcpconn := range testkeys.TCPConnect {
			if tcpconn.Status.Success == true {
				testkeys.TelegramTCPBlocking = false
				break
			}
		}
		for _, roundtrip := range testkeys.Requests {
			if strings.Contains(roundtrip.Request.URL, "web.telegram.org") {
				continue
			}
			if roundtrip.Failure == "" {
				testkeys.TelegramHTTPBlocking = false
				break
			}
		}
		var failure *string
		count := 0
		for _, roundtrip := range testkeys.Requests {
			if !strings.Contains(roundtrip.Request.URL, "web.telegram.org") {
				continue
			}
			count += 1
			if roundtrip.Failure != "" {
				failure = &roundtrip.Failure
				break
			}
		}
		var status string
		if failure != nil {
			testkeys.TelegramWebFailure = failure
			status = "blocked"
		} else if count > 0 {
			status = "ok"
		}
		testkeys.TelegramWebStatus = &status
	}()
	var addresses = []string{
		"149.154.175.50", "149.154.167.51", "149.154.175.100",
		"149.154.167.91", "149.154.171.5",
	}
	var epnts []string
	for _, addr := range addresses {
		epnts = append(epnts, net.JoinHostPort(addr, "80"))
		epnts = append(epnts, net.JoinHostPort(addr, "443"))
	}
	for res := range ootemplate.TCPConnectAsync(ctx, sess.Logger, epnts...) {
		testkeys.TCPConnect = append(testkeys.TCPConnect, res)
	}
	var templates []ootemplate.HTTPRequestTemplate
	for _, addr := range addresses {
		templates = append(templates, ootemplate.HTTPRequestTemplate{
			Method: "POST", URL: "http://" + net.JoinHostPort(addr, "80"),
		})
		// Note: it's intended to connect using `http` on port `443`. I was
		// surprised as well, but this is the spec and using `https` is actually
		// going to lead to I/O timeouts and other failures.
		templates = append(templates, ootemplate.HTTPRequestTemplate{
			Method: "POST", URL: "http://" + net.JoinHostPort(addr, "443"),
		})
	}
	templates = append(templates, ootemplate.HTTPRequestTemplate{
		Method: "GET", URL: "http://web.telegram.org",
	})
	templates = append(templates, ootemplate.HTTPRequestTemplate{
		Method: "GET", URL: "https://web.telegram.org",
	})
	roundtrips, err := ootemplate.HTTPPerformMany(ctx, sess.Logger, templates...)
	if err != nil {
		return err
	}
	testkeys.Requests = roundtrips
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
