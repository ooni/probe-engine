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
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	netxlogger "github.com/ooni/netx/x/logger"
	"github.com/ooni/netx/x/porcelain"
	"github.com/ooni/probe-engine/experiment"
	"github.com/ooni/probe-engine/experiment/handler"
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

func newTestKeys() *TestKeys {
	return &TestKeys{
		TelegramHTTPBlocking: true,
		TelegramTCPBlocking:  true,
		TelegramWebFailure:   nil,
		TelegramWebStatus:    "ok",
	}
}

func (tk *TestKeys) processone(v *urlMeasurements) error {
	if v == nil || v.err != nil {
		return errors.New("passed wrong data to processone")
	}
	r := v.results
	if r == nil {
		return errors.New("passed nil results")
	}
	// update the requests and tcp-connect entries
	tk.Requests = append(
		tk.Requests, oodatamodel.NewRequestList(r)...,
	)
	tk.TCPConnect = append(
		tk.TCPConnect,
		oodatamodel.NewTCPConnectList(r.TestKeys)...,
	)
	// process access points first
	if v.method != "GET" {
		if r.Error == nil {
			tk.TelegramHTTPBlocking = false
			tk.TelegramTCPBlocking = false
			return nil // found successful access point connection
		}
		for _, connect := range r.TestKeys.Connects {
			if connect.Error == nil {
				tk.TelegramTCPBlocking = false
				break // not a connect error meaning we can connect
			}
		}
		return nil
	}
	// now take care of web
	if tk.TelegramWebStatus != "ok" {
		return nil // we already flipped the state
	}
	if r.Error != nil {
		failureString := r.Error.Error()
		tk.TelegramWebStatus = "blocked"
		tk.TelegramWebFailure = &failureString
		return nil
	}
	if r.StatusCode != 200 {
		failureString := "http_request_failed" // MK uses it
		tk.TelegramWebFailure = &failureString
		tk.TelegramWebStatus = "blocked"
		return nil
	}
	title := []byte(`<title>Telegram Web</title>`)
	if bytes.Contains(r.BodySnap, title) == false {
		failureString := "telegram_missing_title_error"
		tk.TelegramWebFailure = &failureString
		tk.TelegramWebStatus = "blocked"
		return nil
	}
	return nil
}

func (tk *TestKeys) processall(m map[string]*urlMeasurements) error {
	for _, v := range m {
		err := tk.processone(v)
		if err != nil {
			return err
		}
	}
	return nil
}

type measurer struct {
	config Config
	do     func(origCtx context.Context,
		config porcelain.HTTPDoConfig) (*porcelain.HTTPDoResults, error)
}

func newMeasurer(config Config) *measurer {
	return &measurer{
		config: config,
		do:     porcelain.HTTPDo,
	}
}

func (m *measurer) measure(
	ctx context.Context,
	sess *session.Session,
	measurement *model.Measurement,
	callbacks handler.Callbacks,
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
	var (
		completed     int64
		receivedBytes int64
		sentBytes     int64
		waitgroup     sync.WaitGroup
	)
	waitgroup.Add(len(urlmeasurements))
	for key := range urlmeasurements {
		go func(key string) {
			defer waitgroup.Done()
			// Avoid making all requests concurrently
			gen := rand.New(rand.NewSource(time.Now().UnixNano()))
			sleeptime := time.Duration(gen.Intn(5000)) * time.Millisecond
			select {
			case <-time.After(sleeptime):
			case <-ctx.Done():
				return
			}
			// No races because each goroutine writes its entry
			entry := urlmeasurements[key]
			entry.results, entry.err = m.do(ctx, porcelain.HTTPDoConfig{
				Handler:   netxlogger.NewHandler(sess.Logger),
				Method:    entry.method,
				URL:       key,
				UserAgent: useragent.Random(),
			})
			if entry.results != nil {
				tk := &entry.results.TestKeys
				atomic.AddInt64(&sentBytes, tk.SentBytes)
				atomic.AddInt64(&receivedBytes, tk.ReceivedBytes)
			}
			sofar := atomic.AddInt64(&completed, 1)
			percentage := float64(sofar) / float64(len(urlmeasurements))
			errstr := "success"
			if entry.err != nil {
				errstr = entry.err.Error()
			}
			callbacks.OnProgress(percentage, fmt.Sprintf(
				"telegram: access %s: %s", key, errstr,
			))
		}(key)
	}
	waitgroup.Wait()
	// fill the measurement entry
	testkeys := newTestKeys()
	measurement.TestKeys = &testkeys
	err := testkeys.processall(urlmeasurements)
	callbacks.OnDataUsage(
		float64(receivedBytes)/1024.0, // downloaded
		float64(sentBytes)/1024.0,     // uploaded
	)
	return err
}

// NewExperiment creates a new experiment.
func NewExperiment(
	sess *session.Session, config Config,
) *experiment.Experiment {
	return experiment.New(sess, testName, testVersion,
		newMeasurer(config).measure)
}
