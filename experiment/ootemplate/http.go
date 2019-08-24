package ootemplate

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"unicode/utf8"

	"github.com/ooni/probe-engine/httpx/httplog"
	"github.com/ooni/probe-engine/httpx/httptracex"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/log"
)

// HTTPTor contains Tor information
type HTTPTor struct {
	ExitIP   *string `json:"exit_ip"`
	ExitName *string `json:"exit_name"`
	IsTor    bool    `json:"is_tor"`
}

// HTTPBody is an HTTP body. We use this helper class to define a custom
// JSON encoder that allows us to choose the proper representation depending
// on whether the Value field is UTF-8 or not.
type HTTPBody struct {
	Value string
}

// MarshalJSON marshal the body to JSON following the OONI spec that says
// that UTF-8 bodies are represened as string and non-UTF-8 bodies are
// instead represented as `{"format":"base64","data":"..."}`.
func (hb HTTPBody) MarshalJSON() ([]byte, error) {
	if utf8.ValidString(hb.Value) {
		return json.Marshal(hb.Value)
	}
	er := make(map[string]string)
	er["format"] = "base64"
	er["data"] = base64.StdEncoding.EncodeToString([]byte(hb.Value))
	return json.Marshal(er)
}

// HTTPRequest contains an HTTP request
type HTTPRequest struct {
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	Tor     HTTPTor           `json:"tor"`
	URL     string            `json:"url"`
	Body    HTTPBody          `json:"body"`
}

// HTTPResponse contains an HTTP response
type HTTPResponse struct {
	Body    HTTPBody          `json:"body"`
	Code    int               `json:"code"`
	Headers map[string]string `json:"headers"`
}

// HTTPRoundTrip contains the measurement of an HTTP round trip.
type HTTPRoundTrip struct {
	Failure  string       `json:"failure"`
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}

// HTTPMeasurer measures several round trips.
type HTTPMeasurer struct {
	RoundTrips   []HTTPRoundTrip
	Mutex        sync.Mutex
	RoundTripper http.RoundTripper
}

func makeroundtrip() HTTPRoundTrip {
	return HTTPRoundTrip{
		Request: HTTPRequest{
			Headers: make(map[string]string),
		},
		Response: HTTPResponse{
			Headers: make(map[string]string),
		},
	}
}

func getheaders(s http.Header) (d map[string]string) {
	d = make(map[string]string)
	for k, v := range s {
		// TODO(bassosimone): this representation of headers loses information
		// and we should instead use a list based representation
		if len(v) > 0 {
			d[k] = v[0]
		}
	}
	return
}

func getbody(s *io.ReadCloser, dest *HTTPBody) (err error) {
	if *s == nil {
		(*dest).Value = ""
		return
	}
	data, err := ioutil.ReadAll(*s)
	if err != nil {
		return
	}
	(*s).Close()
	*s = ioutil.NopCloser(bytes.NewReader(data))
	(*dest).Value = string(data)
	return
}

// RoundTrip performs an HTTP round trip and saves the results.
func (hm *HTTPMeasurer) RoundTrip(req *http.Request) (*http.Response, error) {
	ootrip := makeroundtrip()
	defer func() {
		hm.Mutex.Lock()
		defer hm.Mutex.Unlock()
		hm.RoundTrips = append(hm.RoundTrips, ootrip)
	}()
	ootrip.Request.Headers = getheaders(req.Header)
	ootrip.Request.Method = req.Method
	ootrip.Request.URL = req.URL.String()
	err := getbody(&req.Body, &ootrip.Request.Body)
	if err != nil {
		return nil, err
	}
	resp, err := hm.RoundTripper.RoundTrip(req)
	if err != nil {
		ootrip.Failure = err.Error()
		return resp, err
	}
	ootrip.Response.Code = resp.StatusCode
	ootrip.Response.Headers = getheaders(resp.Header)
	err = getbody(&resp.Body, &ootrip.Response.Body)
	if err != nil {
		ootrip.Failure = err.Error()
		return resp, err
	}
	return resp, nil
}

// NewHTTPClientWithMeasurer is like httpx.NewTracingProxying client except
// that it also returns a measurer where you can see round trips.
//
// The |log| argument is an apex/log.Log compatible logger. The |proxy|
// argument is the function that should be used to setup a proxy (we don't
// need a proxy when measuring, except for circumvention tools). The
// |tlsConfig| argument is an optional TLS configuration with which you
// can, e.g., supply the path to an extrnal CA bundle.
func NewHTTPClientWithMeasurer(
	logger log.Logger, proxy func(req *http.Request) (*url.URL, error),
	tlsConfig *tls.Config,
) (client *http.Client, measurer *HTTPMeasurer) {
	// This creates a transport that contains as transport the measurer in
	// httptracex, which routes events to the |Handler| and uses as a
	// real transport |RoundTripper|. Here |RoundTripper| is a real HTTP
	// transport. The httptracex.Measurer is currently only used for
	// logging via httplog.RoundTripLogger. We have improvements in
	// ooni/probe-engine#26 that will allow us to collect much more low level
	// data than now (e.g. TLS). For now the plan is to just tap into the
	// round-trip to collect the bare minimum required to implement Telegram.
	//
	// TODO(bassosimone): modify the code such that we can take advantage
	// of the httptracex framework to log everything HTTP.
	measurer = &HTTPMeasurer{
		RoundTripper: &httptracex.Measurer{
			RoundTripper: httpx.NewTransport(proxy, tlsConfig),
			Handler: &httplog.RoundTripLogger{
				Logger: logger,
			},
		},
	}
	// This creates an HTTP client where the transport is our measurer
	client = &http.Client{Transport: measurer}
	return
}

// HTTPRequesTemplate is a quick template for setting up an HTTP request
// for performing OONI measurements.
type HTTPRequestTemplate struct {
	Method    string
	URL       string
	UserAgent string
}

// HTTPPerformMany performs the many HTTP requests described by the
// provided template using a custom client and then returns the measurements
// as a list of HTTP round trips. This function returns an error when we
// cannot even prepare a request because you provided us with invalid data,
// otherwise in any other case it will succeed.
func HTTPPerformMany(
	ctx context.Context, logger log.Logger, templates ...HTTPRequestTemplate,
) ([]HTTPRoundTrip, error) {
	var requests []*http.Request
	for _, t := range templates {
		req, err := http.NewRequest(t.Method, t.URL, nil)
		if err != nil {
			return nil, err
		}
		if t.UserAgent == "" {
			// 11.8% as of August 24, 2019 according to
			// https://techblog.willshouse.com/2012/01/03/most-common-user-agents/
			t.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36"
		}
		req.Header.Add("User-Agent", t.UserAgent)
		req = req.WithContext(ctx)
		requests = append(requests, req)
	}
	client, measurer := NewHTTPClientWithMeasurer(logger, nil, nil)
	var waitgroup sync.WaitGroup
	waitgroup.Add(len(requests))
	for _, req := range requests {
		go func(req *http.Request) {
			resp, err := client.Do(req)
			if err == nil {
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
			waitgroup.Done()
		}(req)
	}
	waitgroup.Wait()
	return measurer.RoundTrips, nil
}
