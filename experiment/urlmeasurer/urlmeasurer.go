// Package urlmeasurer contains the URL measurer
package urlmeasurer

import (
	"context"
	"encoding/json"
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

// Output is the output of the URL measurer. We do not emit
// the body as JSON because that would duplicate a large
// piece of data that we already submit to OONI's collector
// using specific keys of the report.
type Output struct {
	Body    []byte `json:"-"`
	Events  []model.Measurement
	Err     error
	logger  log.Logger
	mutex   sync.Mutex
	verbose bool
}

// OnMeasurement handles incoming measurements
func (o *Output) OnMeasurement(meas model.Measurement) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.Events = append(o.Events, meas)
	if o.verbose {
		data, err := json.Marshal(meas)
		if err == nil {
			o.logger.Debugf("%s", data)
		}
	}
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
	req = req.WithContext(ctx)
	resp, out.Err = client.Do(req)
	if out.Err != nil {
		return
	}
	out.Body, out.Err = ioutil.ReadAll(resp.Body)
	resp.Body.Close() // do it synchronously
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
	um.do(ctx, input, client.HTTPClient, out)
	client.Transport.CloseIdleConnections()
	return out
}
