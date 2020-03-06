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

	"github.com/ooni/probe-engine/experiment/httpheader"
	"github.com/ooni/probe-engine/internal/netxlogger"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "telegram"
	testVersion = "0.0.5"
)

// Config contains the experiment config.
type Config struct{}

// TestKeys contains telegram test keys.
type TestKeys struct {
	Agent                string                          `json:"agent"`
	Queries              oonidatamodel.DNSQueriesList    `json:"queries"`
	Requests             oonidatamodel.RequestList       `json:"requests"`
	TCPConnect           oonidatamodel.TCPConnectList    `json:"tcp_connect"`
	TelegramHTTPBlocking bool                            `json:"telegram_http_blocking"`
	TelegramTCPBlocking  bool                            `json:"telegram_tcp_blocking"`
	TelegramWebFailure   *string                         `json:"telegram_web_failure"`
	TelegramWebStatus    string                          `json:"telegram_web_status"`
	TLSHandshakes        oonidatamodel.TLSHandshakesList `json:"tls_handshakes"`
}

type urlMeasurements struct {
	method  string
	results *oonitemplates.HTTPDoResults
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
	if v == nil {
		return errors.New("passed nil data to processone")
	}
	r := v.results
	if r == nil {
		return errors.New("passed nil results")
	}
	tk.Agent = "redirect"
	// update the requests and tcp-connect entries
	tk.Queries = append(
		tk.Queries, oonidatamodel.NewDNSQueriesList(r.TestKeys)...,
	)
	tk.Requests = append(
		tk.Requests, oonidatamodel.NewRequestList(r.TestKeys)...,
	)
	tk.TCPConnect = append(
		tk.TCPConnect,
		oonidatamodel.NewTCPConnectList(r.TestKeys)...,
	)
	tk.TLSHandshakes = append(
		tk.TLSHandshakes,
		oonidatamodel.NewTLSHandshakesList(r.TestKeys)...,
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
		if err := tk.processone(v); err != nil {
			return err
		}
	}
	return nil
}

type measurer struct {
	config Config
}

func (m *measurer) ExperimentName() string {
	return testName
}

func (m *measurer) ExperimentVersion() string {
	return testVersion
}

func (m *measurer) Run(
	ctx context.Context,
	sess model.ExperimentSession,
	measurement *model.Measurement,
	callbacks model.ExperimentCallbacks,
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
		mu            sync.Mutex
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
			mu.Lock()
			entry := urlmeasurements[key]
			mu.Unlock()
			entry.results = oonitemplates.HTTPDo(ctx, oonitemplates.HTTPDoConfig{
				Accept:         httpheader.RandomAccept(),
				AcceptLanguage: httpheader.RandomAcceptLanguage(),
				Beginning:      measurement.MeasurementStartTimeSaved,
				Handler:        netxlogger.NewHandler(sess.Logger()),
				Method:         entry.method,
				URL:            key,
				UserAgent:      httpheader.RandomUserAgent(),
			})
			tk := &entry.results.TestKeys
			atomic.AddInt64(&sentBytes, tk.SentBytes)
			atomic.AddInt64(&receivedBytes, tk.ReceivedBytes)
			sofar := atomic.AddInt64(&completed, 1)
			percentage := float64(sofar) / float64(len(urlmeasurements))
			callbacks.OnProgress(percentage, fmt.Sprintf(
				"telegram: access %s: %s", key, errString(entry.results.Error),
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

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &measurer{config: config}
}

func errString(err error) (s string) {
	s = "success"
	if err != nil {
		s = err.Error()
	}
	return
}
