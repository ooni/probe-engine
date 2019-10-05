// +build !cgo

// Package telegram contains the Telegram network experiment. This file
// in particular is a pure-Go implementation of that.
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
	Queries              []ootemplate.QueryEntry      `json:"queries"`
	Requests             []ootemplate.RequestsEntry   `json:"requests"`
	TCPConnect           []ootemplate.TCPConnectEntry `json:"tcp_connect"`
	TelegramHTTPBlocking bool                         `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool                         `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string                      `json:"telegram_web_failure"`
	TelegramWebStatus    *string                      `json:"telegram_web_status"`
}

func (tk *TestKeys) measureURL(
	ctx context.Context, measurer *urlmeasurer.URLMeasurer,
	method, URL string,
) {
	output := measurer.Do(ctx, urlmeasurer.Input{
		Method: method,
		URL:    URL,
	})
	tk.Queries = append(tk.Queries, ootemplate.Queries(
		ctx, measurer.DNSNetwork, measurer.DNSAddress, output.Events,
	)...)
	tk.TCPConnect = append(tk.TCPConnect, ootemplate.TCPConnect(output.Events)...)
	tk.Requests = append(tk.Requests, ootemplate.Requests(output.Events)...)
}

func (tk *TestKeys) measureAll(ctx context.Context, sess *session.Session) {
	measurer := &urlmeasurer.URLMeasurer{
		CABundlePath: sess.CABundlePath(),
		DNSNetwork:   "system",
		Logger:       sess.Logger,
	}
	var addresses = []string{
		"149.154.175.50", "149.154.167.51", "149.154.175.100",
		"149.154.167.91", "149.154.171.5",
	}
	for _, addr := range addresses {
		tk.measureURL(ctx, measurer, "POST", "http://"+net.JoinHostPort(addr, "80"))
		// Note: it's intended to connect using `http` on port `443`. I was
		// surprised as well, but this is the spec and using `https` is actually
		// going to lead to I/O timeouts and other failures.
		tk.measureURL(ctx, measurer, "POST", "http://"+net.JoinHostPort(addr, "443"))
	}
	tk.measureURL(ctx, measurer, "GET", "http://web.telegram.org/")
	tk.measureURL(ctx, measurer, "GET", "https://web.telegram.org/")
}

func (tk *TestKeys) analyze() {
	// "If all TCP connections on ports 80 and 443 to Telegram's access point
	// IPs fail we consider Telegram to be blocked"
	for _, tcpconn := range tk.TCPConnect {
		if tcpconn.Status.Success == true {
			tk.TelegramTCPBlocking = false
			break
		}
	}
	// "If at least an HTTP request [to an access point] returns back a
	// response, we consider Telegram to not be blocked"
	for _, roundtrip := range tk.Requests {
		if strings.Contains(roundtrip.Request.URL, "web.telegram.org") {
			// Here we exclude the web endpoints because they are
			// processed below to compute web stats.
			continue
		}
		if roundtrip.Failure == nil {
			tk.TelegramHTTPBlocking = false
			break
		}
	}
	// "If the HTTP(S) requests fail or the HTML <title> tag text is not
	// `Telegram Web` we consider the web version of Telegram to be blocked."
	var failure *string
	count := 0
	for _, roundtrip := range tk.Requests {
		if !strings.Contains(roundtrip.Request.URL, "web.telegram.org") {
			// Here we exclude the non-web-endpoints because we have
			// already processed them above to compute other stats
			continue
		}
		count += 1
		if roundtrip.Failure != nil {
			failure = roundtrip.Failure
			break
		}
		if strings.Contains(
			roundtrip.Response.Body.Value, "<title>Telegram Web</title>",
		) == false {
			v := `Page title is not "Telegram Web"`
			failure = &v
			break
		}
	}
	var status string
	if failure != nil {
		tk.TelegramWebFailure = failure
		status = "blocked"
	} else if count > 0 {
		status = "ok"
	}
	tk.TelegramWebStatus = &status
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
	testkeys.measureAll(ctx, sess)
	testkeys.analyze()
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
