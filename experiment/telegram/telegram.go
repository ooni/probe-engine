// Package telegram contains the Telegram network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-020-telegram.md.
package telegram

import (
	"context"
	"strings"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

const (
	testName    = "telegram"
	testVersion = "0.1.1"
)

// Config contains the telegram experiment config.
type Config struct{}

// TestKeys contains telegram test keys.
type TestKeys struct {
	urlgetter.TestKeys
	TelegramHTTPBlocking bool    `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool    `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string `json:"telegram_web_failure"`
	TelegramWebStatus    string  `json:"telegram_web_status"`
}

// NewTestKeys creates new telegram TestKeys.
func NewTestKeys() *TestKeys {
	return &TestKeys{
		TelegramHTTPBlocking: true,
		TelegramTCPBlocking:  true,
		TelegramWebFailure:   nil,
		TelegramWebStatus:    "ok",
	}
}

// Update updates the TestKeys using the given MultiOutput result.
func (tk *TestKeys) Update(v urlgetter.MultiOutput) {
	// update the easy to update entries first
	tk.NetworkEvents = append(tk.NetworkEvents, v.TestKeys.NetworkEvents...)
	tk.Queries = append(tk.Queries, v.TestKeys.Queries...)
	tk.Requests = append(tk.Requests, v.TestKeys.Requests...)
	tk.TCPConnect = append(tk.TCPConnect, v.TestKeys.TCPConnect...)
	tk.TLSHandshakes = append(tk.TLSHandshakes, v.TestKeys.TLSHandshakes...)
	// then process access points
	if v.Input.Config.Method != "GET" {
		if v.TestKeys.Failure == nil {
			tk.TelegramHTTPBlocking = false
			tk.TelegramTCPBlocking = false
			return // found successful access point connection
		}
		if v.TestKeys.FailedOperation == nil || *v.TestKeys.FailedOperation != modelx.ConnectOperation {
			tk.TelegramTCPBlocking = false
		}
		return
	}
	// now take care of web
	if tk.TelegramWebStatus != "ok" {
		return // we already flipped the state
	}
	if v.TestKeys.Failure != nil {
		tk.TelegramWebStatus = "blocked"
		tk.TelegramWebFailure = v.TestKeys.Failure
		return
	}
	if v.TestKeys.HTTPResponseStatus != 200 {
		failureString := "http_request_failed" // MK uses it
		tk.TelegramWebFailure = &failureString
		tk.TelegramWebStatus = "blocked"
		return
	}
	title := `<title>Telegram Web</title>`
	if strings.Contains(v.TestKeys.HTTPResponseBody, title) == false {
		failureString := "telegram_missing_title_error"
		tk.TelegramWebFailure = &failureString
		tk.TelegramWebStatus = "blocked"
		return
	}
	return
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

func (m measurer) Run(ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	urlgetter.RegisterExtensions(measurement)
	inputs := []urlgetter.MultiInput{
		{Target: "http://149.154.175.50/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.167.51/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.175.100/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.167.91/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.171.5/", Config: urlgetter.Config{Method: "POST"}},

		{Target: "http://149.154.175.50:443/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.167.51:443/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.175.100:443/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.167.91:443/", Config: urlgetter.Config{Method: "POST"}},
		{Target: "http://149.154.171.5:443/", Config: urlgetter.Config{Method: "POST"}},

		{Target: "http://web.telegram.org/", Config: urlgetter.Config{Method: "GET"}},
		{Target: "https://web.telegram.org/", Config: urlgetter.Config{Method: "GET"}},
	}
	multi := urlgetter.Multi{Begin: time.Now(), Session: sess}
	testkeys := NewTestKeys()
	testkeys.Agent = "redirect"
	measurement.TestKeys = testkeys
	for entry := range multi.Collect(ctx, inputs, "telegram", callbacks) {
		testkeys.Update(entry)
	}
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
