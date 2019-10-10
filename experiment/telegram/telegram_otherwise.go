// +build !cgo

// Package telegram contains the Telegram network experiment. This file
// in particular is a pure-Go implementation of that.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-020-telegram.md.
package telegram

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/netx/handlers"
	"github.com/ooni/netx/httpx"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	testName    = "telegram"
	testVersion = "0.0.3"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains telegram test keys.
type TestKeys struct {
	// TODO(bassosimone):
	//
	// 1. we don't fill Telegram{HTTP,TCP}Blocking for now
	//
	// 2. we need to fill keys for the parent data format, and
	// specifically `tcp_connect` and `requests`
	//
	// Both issues will be addressed later when we will
	// start processing ooni/netx events.
	TelegramHTTPBlocking bool    `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool    `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string `json:"telegram_web_failure"`
	TelegramWebStatus    string  `json:"telegram_web_status"`
}

type measurer struct {
	client *http.Client
	logger log.Logger
	tk     TestKeys
}

func (m *measurer) request(
	ctx context.Context, method, scheme, address, port string,
) (*http.Response, error) {
	URL := url.URL{}
	URL.Scheme = scheme
	if port != "" {
		URL.Host = net.JoinHostPort(address, port)
	} else {
		URL.Host = address
	}
	m.logger.Debugf("telegram: %s %s...", method, URL.String())
	req, err := http.NewRequest(method, URL.String(), nil)
	if err != nil {
		return nil, err
	}
	return m.client.Do(req)
}

func (m *measurer) measureDC(ctx context.Context) {
	var addresses = []string{
		"149.154.175.50", "149.154.167.51", "149.154.175.100",
		"149.154.167.91", "149.154.171.5",
	}
	for _, addr := range addresses {
		for _, port := range []string{"80", "443"} {
			const alwaysHTTP = "http" // note: it's intended to use HTTP with 443
			resp, err := m.request(ctx, "POST", alwaysHTTP, addr, port)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			//
			// note: we expect server to return 501 Not Implemented, but relying on
			// this isn't required by the spec and won't be super robust
			//
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}
		}
	}
}

func (m *measurer) measureWebWithScheme(ctx context.Context, scheme string) error {
	// "If the HTTP(S) requests fail or the HTML <title> tag text is not
	// `Telegram Web` we consider the web version of Telegram to be blocked."
	resp, err := m.request(ctx, "GET", scheme, "web.telegram.org", "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("HTTP request failed")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	title := []byte(`<title>Telegram Web</title>`)
	if bytes.Contains(data, title) == false {
		return errors.New(`Page title is not "Telegram Web"`)
	}
	return nil
}

func (m *measurer) measureWeb(ctx context.Context) {
	// "If the HTTP(S) requests fail or the HTML <title> tag text is not
	// `Telegram Web` we consider the web version of Telegram to be blocked."
	m.tk.TelegramWebStatus = "ok"
	errHTTP := m.measureWebWithScheme(ctx, "http")
	errHTTPS := m.measureWebWithScheme(ctx, "https")
	if errHTTP != nil {
		s := errHTTP.Error()
		m.tk.TelegramWebFailure = &s
		m.tk.TelegramWebStatus = "failure"
	} else if errHTTPS != nil {
		s := errHTTPS.Error()
		m.tk.TelegramWebFailure = &s
		m.tk.TelegramWebStatus = "failure"
	}
}

func (m *measurer) analyze() {
	// TODO(bassosimone): this is where we need to process
	// netx events and fill more keys in the results
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	client := httpx.NewClient(handlers.StdoutHandler)
	measurer := &measurer{
		client: client.HTTPClient,
		logger: sess.Logger,
	}
	measurement.TestKeys = &measurer.tk
	client.SetCABundle(sess.CABundlePath())
	measurer.measureDC(ctx)
	measurer.measureWeb(ctx)
	measurer.analyze()
	client.Transport.CloseIdleConnections()
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
