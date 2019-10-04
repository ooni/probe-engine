// Package urlmeasurer contains the URL measurer
package urlmeasurer

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/ooni/netx/httpx"
	"github.com/ooni/netx/model"
	"github.com/ooni/probe-engine/log"
)

// URLMeasurer is a measurer for URLs.
type URLMeasurer struct {
	CABundlePath string
	DNSNetwork   string
	DNSAddress   string
	Logger       log.Logger
	Verbose      bool
}

// Input is the input to the URL measurer.
type Input struct {
	Method string
	URL    string
}

// Output is the output of the URL measurer.
type Output struct {
	Events       [][]model.Measurement
	Err          error
	current      []model.Measurement
	logger       log.Logger
	roundTripper http.RoundTripper
	mutex        sync.Mutex
	verbose      bool
}

// OnMeasurement handles incoming measurements
func (o *Output) OnMeasurement(meas model.Measurement) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.current = append(o.current, meas)
	if o.verbose {
		o.logger.Debugf("%+v", meas)
	}
}

// RoundTrip implements the http.RoundTripper interface
func (o *Output) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := o.roundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// Fully read the response body so to see all the round trip events
	// and have all of them inside of the o.current buffer.
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	// Move events of the current round trip into the archive so that
	// all the events we have are organized by round trip.
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.Events = append(o.Events, o.current)
	o.current = nil
	return resp, err
}

func (um *URLMeasurer) do(
	ctx context.Context, input Input, client *http.Client, out *Output,
) {
	var (
		req  *http.Request
		resp *http.Response
	)
	req, out.Err = http.NewRequest(input.Method, input.URL, nil)
	if out.Err != nil {
		return
	}
	// 11.8% as of August 24, 2019 according to
	// https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
	)
	req = req.WithContext(ctx)
	resp, out.Err = client.Do(req)
	if out.Err != nil {
		return
	}
	defer resp.Body.Close()
	_, out.Err = ioutil.ReadAll(resp.Body)
	return
}

// Do measures the required URL and returns the output.
func (um *URLMeasurer) Do(ctx context.Context, input Input) *Output {
	out := &Output{
		logger:  um.Logger,
		verbose: um.Verbose,
	}
	client := httpx.NewClient(out)
	client.SetCABundle(um.CABundlePath)
	out.Err = client.ConfigureDNS(um.DNSNetwork, um.DNSAddress)
	if out.Err != nil {
		return out
	}
	// replace the round tripper with our round tripper than ensures
	// that events are conveniently divided by round trip.
	out.roundTripper = client.HTTPClient.Transport
	client.HTTPClient.Transport = out
	um.do(ctx, input, client.HTTPClient, out)
	client.Transport.CloseIdleConnections()
	return out
}
