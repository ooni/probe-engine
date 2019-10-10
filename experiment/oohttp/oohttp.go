// Package oohttp contains OONI's HTTP client.
package oohttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ooni/netx/httpx"
	"github.com/ooni/netx/model"
)

// TODO(bassosimone): this user-agent solution is temporary and we
// should instead select one among many user agents. We should open
// an issue before merging this PR to address this defect.

// 11.8% as of August 24, 2019 according to https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"

type roundTripMeasurements struct {
	measurements []model.Measurement
	mutex        sync.Mutex
}

func (rtm *roundTripMeasurements) OnMeasurement(m model.Measurement) {
	rtm.mutex.Lock()
	defer rtm.mutex.Unlock()
	rtm.measurements = append(rtm.measurements, m)
}

// MeasuringClient gets you an *http.Client configured to perform
// ooni/netx measurements during round trips.
type MeasuringClient struct {
	client       *http.Client
	close        func() error
	measurements [][]model.Measurement
	mutex        sync.Mutex
	rtm          *roundTripMeasurements
	transport    http.RoundTripper
}

// Config contains the configuration
type Config struct {
	// CABundlePath is the path of the CA bundle to use. If empty we
	// will be using the system default CA bundle.
	CABundlePath string
}

// NewMeasuringClient creates a new MeasuringClient instance.
func NewMeasuringClient(config Config) *MeasuringClient {
	mc := new(MeasuringClient)
	mc.client = &http.Client{
		Transport: mc,
	}
	mc.rtm = new(roundTripMeasurements)
	transport := httpx.NewTransport(time.Now(), mc.rtm)
	transport.SetCABundle(config.CABundlePath)
	mc.transport = transport
	mc.close = func() error {
		transport.CloseIdleConnections()
		return nil
	}
	return mc
}

// HTTPClient returns the *http.Client you should be using.
func (mc *MeasuringClient) HTTPClient() *http.Client {
	return mc.client
}

// PopMeasurementsByRoundTrip returns the ooni/netx measurements organized
// by round trip and clears the internal measurements cache.
func (mc *MeasuringClient) PopMeasurementsByRoundTrip() [][]model.Measurement {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	out := mc.measurements
	mc.measurements = nil
	return out
}

// RoundTrip performs a RoundTrip.
func (mc *MeasuringClient) RoundTrip(req *http.Request) (*http.Response, error) {
	// Make sure we have a browser user agent for measurements.
	req.Header.Set("User-Agent", userAgent)
	resp, err := mc.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// Fully read the response body so to see all the round trip events
	// and have all of them inside of the c.current buffer.
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	// Move events of the current round trip into the archive so that
	// all the events we have are organized by round trip.
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.measurements = append(mc.measurements, mc.rtm.measurements)
	mc.rtm.measurements = nil
	return resp, err
}

// Close closes the resources we may have openned
func (mc *MeasuringClient) Close() error {
	return mc.close()
}
