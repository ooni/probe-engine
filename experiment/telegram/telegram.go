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
	"math/rand"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/netxlogger"
	"github.com/ooni/probe-engine/experiment/oodatamodel"
	"github.com/ooni/probe-engine/experiment/useragent"
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
	Requests             oodatamodel.RequestList    `json:"requests"`
	TCPConnect           oodatamodel.TCPConnectList `json:"tcp_connect"`
	TelegramHTTPBlocking bool                       `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool                       `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string                    `json:"telegram_web_failure"`
	TelegramWebStatus    string                     `json:"telegram_web_status"`
}

type urlMeasurements struct {
	method  string
	err     error
	results *porcelain.HTTPDoResults
}

func measure(
	ctx context.Context, sess *session.Session, measurement *model.Measurement,
	callbacks handler.Callbacks, config Config,
) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// setup data container
	var urlmeasurements = map[string]*urlMeasurements{
		"http://149.154.175.50/":  &urlMeasurements{method: "POST"},
		"http://149.154.167.51/":  &urlMeasurements{method: "POST"},
		"http://149.154.175.100/": &urlMeasurements{method: "POST"},
		"http://149.154.167.91/":  &urlMeasurements{method: "POST"},
		"http://149.154.171.5/":   &urlMeasurements{method: "POST"},

		"http://149.154.175.50:443/":  &urlMeasurements{method: "POST"},
		"http://149.154.167.51:443/":  &urlMeasurements{method: "POST"},
		"http://149.154.175.100:443/": &urlMeasurements{method: "POST"},
		"http://149.154.167.91:443/":  &urlMeasurements{method: "POST"},
		"http://149.154.171.5:443/":   &urlMeasurements{method: "POST"},

		"http://web.telegram.org/":  &urlMeasurements{method: "GET"},
		"https://web.telegram.org/": &urlMeasurements{method: "GET"},
	}

	// run all measurements in parallel
	var waitgroup sync.WaitGroup
	waitgroup.Add(len(urlmeasurements))
	for key := range urlmeasurements {
		go func(key string) {
			// Avoid making all requests concurrently
			gen := rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Duration(gen.Intn(5000)) * time.Millisecond)
			// No races because each goroutine writes its entry
			entry := urlmeasurements[key]
			entry.results, entry.err = porcelain.HTTPDo(ctx, porcelain.HTTPDoConfig{
				Handler:   netxlogger.New(log.Log),
				Method:    entry.method,
				URL:       key,
				UserAgent: useragent.Random(),
			})
			waitgroup.Done()
		}(key)
	}
	waitgroup.Wait()

	// fill the measurement entry
	testkeys := &TestKeys{
		TelegramHTTPBlocking: true,
		TelegramTCPBlocking:  true,
		TelegramWebFailure:   nil,
		TelegramWebStatus:    "ok",
	}
	measurement.TestKeys = &testkeys
	for _, v := range urlmeasurements {
		if v.err != nil {
			return errors.New("passed wrong data to netx/porcelain")
		}
		r := v.results
		// update the requests and tcp-connect entries
		testkeys.Requests = append(
			testkeys.Requests, oodatamodel.NewRequestList(r)...,
		)
		testkeys.TCPConnect = append(
			testkeys.TCPConnect,
			oodatamodel.NewTCPConnectList(r.TestKeys)...,
		)
		// process access points first
		if v.method != "GET" {
			if r.Error == nil {
				testkeys.TelegramHTTPBlocking = false
				testkeys.TelegramTCPBlocking = false
				continue // found successful access point connection
			}
			for _, connect := range r.TestKeys.Connects {
				if connect.Error == nil {
					testkeys.TelegramTCPBlocking = false
					break // not a connect error meaning we can connect
				}
			}
			continue
		}
		// now take care of web
		if testkeys.TelegramWebStatus != "ok" {
			continue // we already flipped the state
		}
		if r.Error != nil {
			failureString := r.Error.Error()
			testkeys.TelegramWebStatus = "blocked"
			testkeys.TelegramWebFailure = &failureString
			continue
		}
		if r.StatusCode != 200 {
			failureString := "http_request_failed" // MK uses it
			testkeys.TelegramWebFailure = &failureString
			testkeys.TelegramWebStatus = "blocked"
			continue
		}
		title := []byte(`<title>Telegram Web</title>`)
		if bytes.Contains(r.Body, title) == false {
			failureString := "telegram_missing_title_error"
			testkeys.TelegramWebFailure = &failureString
			testkeys.TelegramWebStatus = "blocked"
			continue
		}
	}

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
