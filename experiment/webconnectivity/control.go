package webconnectivity

import (
	"context"
	"net"
	"sort"
	"strconv"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/geoiplookup/mmdblookup"
	"github.com/ooni/probe-engine/internal/httpx"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/modelx"
)

// ControlRequest is the request that we send to the control
type ControlRequest struct {
	HTTPRequest        string              `json:"http_request"`
	HTTPRequestHeaders map[string][]string `json:"http_request_headers"`
	TCPConnect         []string            `json:"tcp_connect"`
}

// NewControlRequest creates a new control request from the target
// URL and the result obtained by measuring it.
func NewControlRequest(
	target model.MeasurementTarget, result urlgetter.TestKeys) (out ControlRequest) {
	out.HTTPRequest = string(target)
	out.HTTPRequestHeaders = make(map[string][]string)
	if len(result.Requests) >= 1 { // defensive
		// Only these headers are accepted by the control service. We use the
		// canonical header keys, which is what Go enforces.
		for _, key := range []string{"User-Agent", "Accept", "Accept-Language"} {
			// We never set multi line headers so using the map instead of the list
			// is totally fine and saves us some extra lines of code. Also, we never
			// put binary values in headers, so we can just use value.Value.
			if value, ok := result.Requests[0].Request.Headers[key]; ok {
				out.HTTPRequestHeaders[key] = []string{value.Value}
			}
		}
	}
	// Just in case we have multiple endpoints/attempts, reduce them:
	out.TCPConnect = []string{}
	epnts := make(map[string]int)
	for _, entry := range result.TCPConnect {
		epnts[net.JoinHostPort(entry.IP, strconv.Itoa(entry.Port))]++
	}
	for key := range epnts {
		out.TCPConnect = append(out.TCPConnect, key)
	}
	sort.Slice(out.TCPConnect, func(i, j int) bool { // stable output wrt map iteration
		return out.TCPConnect[i] < out.TCPConnect[j]
	})
	return
}

// ControlTCPConnectResult is the result of the TCP connect
// attempt performed by the control vantage point.
type ControlTCPConnectResult struct {
	Status  bool    `json:"status"`
	Failure *string `json:"failure"`
}

// ControlHTTPRequestResult is the result of the HTTP request
// performed by the control vantage point.
type ControlHTTPRequestResult struct {
	BodyLength int64             `json:"body_length"`
	Failure    *string           `json:"failure"`
	Title      string            `json:"title"`
	Headers    map[string]string `json:"headers"`
	StatusCode int64             `json:"status_code"`
}

// ControlDNSResult is the result of the DNS lookup
// performed by the control vantage point.
type ControlDNSResult struct {
	Failure *string  `json:"failure"`
	Addrs   []string `json:"addrs"`
	ASNs    []int64  `json:"x_asns"` // not in spec
}

// ControlResponse is the response from the control service.
type ControlResponse struct {
	TCPConnect  map[string]ControlTCPConnectResult `json:"tcp_connect"`
	HTTPRequest ControlHTTPRequestResult           `json:"http_request"`
	DNS         ControlDNSResult                   `json:"dns"`
}

// Control performs the control request and returns the response.
func Control(
	ctx context.Context, sess model.ExperimentSession,
	thAddr string, creq ControlRequest) (out ControlResponse, err error) {
	clnt := httpx.Client{
		BaseURL:    thAddr,
		HTTPClient: sess.DefaultHTTPClient(),
		Logger:     sess.Logger(),
	}
	// make sure error is wrapped
	err = errorx.SafeErrWrapperBuilder{
		Error:     clnt.PostJSON(ctx, "/", creq, &out),
		Operation: modelx.TopLevelOperation,
	}.MaybeBuild()
	(&out.DNS).FillASNs(sess)
	return
}

// FillASNs fills the ASNs array of ControlDNSResult. For each Addr inside
// of the ControlDNSResult structure, we obtain the corresponding ASN.
//
// This is very useful to know what ASNs were the IP addresses returned by
// the control according to the probe's ASN database.
func (dns *ControlDNSResult) FillASNs(sess model.ExperimentSession) {
	dns.ASNs = []int64{}
	for _, ip := range dns.Addrs {
		// TODO(bassosimone): this would be more efficient if we'd open just
		// once the database and then reuse it for every address.
		asn, _, _ := mmdblookup.ASN(sess.ASNDatabasePath(), ip)
		dns.ASNs = append(dns.ASNs, int64(asn))
	}
}
