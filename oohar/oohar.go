// Package oohar implements the OONI HAR format.
package oohar

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/httpx/minihar"
	"github.com/ooni/probe-engine/httpx/tlsx"
)

// HeaderInfo contains information about a header.
type HeaderInfo struct {
	// Name is the header name.
	Name string `json:"name"`

	// Value is the header value.
	Value string `json:"value"`

	// Comment is a comment about the header.
	Comment string `json:"comment"`
}

// PostDataInfo contains info about the request body.
type PostDataInfo struct {
	// MIMEType contains the body MIME type.
	MIMEType string `json:"mimeType"`

	// Params contains a list of posted params.
	Params []interface{} `json:"params"`

	// Text is the request body encoded as base64.
	Text []byte `json:"text"`
}

// RequestInfo contains detailed info about the request.
type RequestInfo struct {
	// Method is the request method
	Method string `json:"method"`

	// URL is the absolute URL of the request without fragments.
	URL string `json:"url"`

	// HTTPVersion is the HTTP version.
	HTTPVersion string `json:"httpVersion"`

	// Cookies is a list of cookie objects.
	Cookies []interface{} `json:"cookies"`

	// Headers is a list of header objects.
	Headers []HeaderInfo `json:"headers"`

	// QueryString contains information about the query string.
	QueryString []interface{} `json:"queryString"`

	// PostDaya contains information about the request body.
	PostData *PostDataInfo `json:"postData,omitempty"`

	// HeadersSize is the size of headers or -1 if not available.
	HeadersSize int64 `json:"headersSize"`

	// BodySize is the size of the body or -1 if not available.
	BodySize int64 `json:"bodySize"`
}

// ContentInfo contains information about the response body.
type ContentInfo struct {
	// Size is the response body size.
	Size int64 `json:"size"`

	// MIMEType contains the body MIME type.
	MIMEType string `json:"mimeType"`

	// Text is the request body encoded as base64.
	Text []byte `json:"text"`
}

// ResponseInfo contains detailed info about the response.
type ResponseInfo struct {
	// Status is the response status.
	Status string `json:"status"`

	// StatusText is the response status text.
	StatusText string `json:"statusText"`

	// HTTPVersion is the HTTP version.
	HTTPVersion string `json:"httpVersion"`

	// Cookies is a list of cookie objects.
	Cookies []interface{} `json:"cookies"`

	// Headers is a list of header objects.
	Headers []HeaderInfo `json:"headers"`

	// Content contains info about the response body.
	Content ContentInfo `json:"content"`

	// RedirectURL is the redirect URL, if any.
	RedirectURL string `json:"redirectURL"`

	// HeadersSize is the size of headers or -1 if not available.
	HeadersSize int64 `json:"headersSize"`

	// BodySize is the size of the body or -1 if not available.
	BodySize int64 `json:"bodySize"`
}

// TimingInfo contains timing info about the round-trip.
type TimingInfo struct {
	// DNS is the time spent resolving the host name or -1 if not available.
	DNS int64 `json:"dns"`

	// Connect is the time spent connecting or -1 if not available.
	Connect int64 `json:"connect"`

	// Send is the time spent sending the request or -1 if not available.
	Send int64 `json:"send"`

	// Wait is the time spent waiting for the response or -1 if not available.
	Wait int64 `json:"wait"`

	// Receive is the time spent receiving the response or -1 if not available.
	Receive int64 `json:"receive"`

	// SSL is the time spent TLS-handshaking or -1 if not available.
	SSL int64 `json:"ssl"`
}

// XDNSInfo contains extended DNS info
type XDNSInfo struct {
	// StartedDateTime is the time when this request started using ISO8601.
	StartedDateTime time.Time `json:"startedDateTime"`

	// Time is the time elapsed in this request
	Time int64 `json:"time"`

	// HostName is the hostname we wanted to resolve
	HostName string `json:"hostName"`

	// Addresses contains the addresses we resolved
	Addresses []string `json:"addresses"`

	// Failure is the result of the DNS
	Failure string `json:"failure"`
}

// XConnectInfo contains extended connect info
type XConnectInfo struct {
	// StartedDateTime is the time when this request started using ISO8601.
	StartedDateTime time.Time `json:"startedDateTime"`

	// Time is the time elapsed in connecting
	Time int64 `json:"time"`

	// Endpoint is the endpoint we wanted to use
	Endpoint string `json:"endpoint"`

	// Failure is the result of the DNS
	Failure string `json:"failure"`
}

// XTLSInfo contains extended TLS info
type XTLSInfo struct {
	// StartDateTime is when we started the TLS handshake.
	StartDateTime time.Time `json:"startDateTime"`

	// Time is the time elapsed in the handshake
	Time int64 `json:"time"`

	// NegotiatedProtocol is the ALPN negotiated protocol
	NegotiatedProtocol string `json:"negotiatedProtocol"`

	// Version is the version of TLS we're using
	Version string `json:"version"`

	// CiperSuite is the cipher suite we use
	CipherSuite string `json:"cipherSuite"`

	// Certs contains the certificates
	Certs []string `json:"certs"`

	// Failure is the result of the TLS handshake.
	Failure string
}

// Entry is a tracker request.
type Entry struct {
	// StartedDateTime is the time when this request started using ISO8601.
	StartedDateTime time.Time `json:"startedDateTime"`

	// Time is the total request time in millisecond.
	Time int64 `json:"time"`

	// Request contains detailed info about the request.
	Request RequestInfo `json:"request"`

	// Response contains detailed info about the response.
	Response ResponseInfo `json:"response"`

	// Cache contains detailed info about cache usage.
	Cache interface{} `json:"cache"`

	// Timings contains timing info about the round-trip.
	Timings TimingInfo `json:"timings"`

	// DNS contains info on DNS resolution.
	DNS []*XDNSInfo `json:"_dns"`

	// Connect contains info on the TCP connect.
	Connect []*XConnectInfo `json:"_connect"`

	// TLS contains info on TLS.
	TLS *XTLSInfo `json:"_tls"`
}

// CreatorInfo contains info on the application that created this log.
type CreatorInfo struct {
	// Name is the name of the application
	Name string `json:"name"`

	// Version is the version of the application
	Version string `json:"version"`

	// Comment is a comment to the creator
	Comment string `json:"comment"`
}

// Log is the oohar log.
type Log struct {
	// Version is the version of this HAR log.
	Version string `json:"version"`

	// Creator is the application that created the log.
	Creator CreatorInfo `json:"creator"`

	// Entries contains the tracker requests.
	Entries []*Entry `json:"entries"`
}

// HAR is the structure of a HAR file
type HAR struct {
	// Log contains the HAR log
	Log Log `json:"log"`
}

func (e *Entry) fillStartedDateTime(rts *minihar.RoundTripSaver) {
	e.StartedDateTime = rts.RoundTripStartTime
}

func elapsed(start, end time.Time) int64 {
	if end.After(start) {
		return int64(end.Sub(start) / time.Millisecond)
	}
	return -1
}

func (e *Entry) fillTime(rts *minihar.RoundTripSaver) {
	e.Time = elapsed(rts.RoundTripStartTime, rts.ResponseBodyCloseTime)
}

func makeHeaders(headers http.Header) (info []HeaderInfo) {
	for key, values := range headers {
		for _, value := range values {
			info = append(info, HeaderInfo{
				Name:  key,
				Value: value,
			})
		}
	}
	return
}

func makeBodySize(info []minihar.ReadInfo) int64 {
	count := int64(-1)
	for _, e := range info {
		if e.Count > 0 {
			count += int64(e.Count)
		}
	}
	return count
}

func (e *Entry) fillRequest(rts *minihar.RoundTripSaver) {
	e.Request.Method = rts.RequestMethod
	e.Request.URL = rts.RequestURL.String() // TODO(bassosimone): no fragment
	e.Request.HTTPVersion = rts.RequestProto
	e.Request.Headers = makeHeaders(rts.RequestHeaders)
	e.Request.PostData = nil
	e.Request.HeadersSize = -1
	e.Request.BodySize = makeBodySize(rts.RequestBodyReadInfo)
}

func (e *Entry) fillResponse(rts *minihar.RoundTripSaver) {
	e.Response.Status = fmt.Sprintf("%d", rts.ResponseStatusCode)
	e.Response.StatusText = http.StatusText(rts.ResponseStatusCode)
	e.Response.HTTPVersion = rts.ResponseProto
	e.Response.Headers = makeHeaders(rts.ResponseHeaders)
	e.Response.Content = ContentInfo{
		Size:     -1,
		MIMEType: "",
		Text:     []byte{},
	}
	e.Response.RedirectURL = rts.ResponseHeaders.Get("Location")
	e.Response.HeadersSize = -1
	e.Response.BodySize = makeBodySize(rts.ResponseBodyReadInfo)
}

func dnsTime(rts *minihar.RoundTripSaver) int64 {
	size := len(rts.DNS)
	if size <= 0 {
		return -1
	}
	return elapsed(rts.DNS[size-1].StartTime, rts.DNS[size-1].EndTime)
}

func connectTime(rts *minihar.RoundTripSaver) int64 {
	entry, ok := rts.Connects[rts.ConnectionEndpoint]
	if !ok {
		return -1
	}
	size := len(entry)
	if size <= 0 {
		return -1
	}
	return elapsed(entry[size-1].StartTime, entry[size-1].EndTime)
}

func (e *Entry) fillTimings(rts *minihar.RoundTripSaver) {
	e.Timings.DNS = dnsTime(rts)
	e.Timings.Connect = connectTime(rts)
	e.Timings.Send = elapsed(rts.ConnectionReadyTime, rts.RequestSentTime)
	e.Timings.Wait = elapsed(rts.RequestSentTime, rts.ResponseFirstByteTime)
	e.Timings.Receive = elapsed(
		rts.ResponseFirstByteTime, rts.ResponseBodyCloseTime,
	)
	e.Timings.SSL = elapsed(rts.TLSHandshakeStartTime, rts.TLSHandshakeEndTime)
}

func strerror(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (e *Entry) fillDNS(rts *minihar.RoundTripSaver) {
	for _, info := range rts.DNS {
		xdns := &XDNSInfo{
			HostName:        info.HostName,
			StartedDateTime: info.StartTime,
		}
		for _, addr := range info.Addresses {
			xdns.Addresses = append(xdns.Addresses, addr.String())
		}
		xdns.Failure = strerror(info.Error)
		xdns.Time = elapsed(info.StartTime, info.EndTime)
		e.DNS = append(e.DNS, xdns)
	}
}

func (e *Entry) fillConnect(rts *minihar.RoundTripSaver) {
	for epnt, infos := range rts.Connects {
		for _, info := range infos {
			xconnect := &XConnectInfo{
				Endpoint:        epnt,
				StartedDateTime: info.StartTime,
			}
			xconnect.Failure = strerror(info.Error)
			xconnect.Time = elapsed(info.StartTime, info.EndTime)
			e.Connect = append(e.Connect, xconnect)
		}
	}
}

func (e *Entry) fillTLS(rts *minihar.RoundTripSaver) {
	if rts.TLSHandshakeStartTime.After(rts.RoundTripStartTime) {
		e.TLS = &XTLSInfo{
			StartDateTime: rts.TLSHandshakeStartTime,
			Time: elapsed(rts.TLSHandshakeStartTime, rts.TLSHandshakeEndTime),
			NegotiatedProtocol: rts.TLSConnectionState.NegotiatedProtocol,
			Version: tlsx.TLSVersionString[rts.TLSConnectionState.Version],
			CipherSuite: tlsx.TLSCipherSuiteString[rts.TLSConnectionState.CipherSuite],
			Failure: strerror(rts.TLSHandshakeError),
		}
		for _, cert := range rts.TLSConnectionState.PeerCertificates {
			e.TLS.Certs = append(e.TLS.Certs, tlsx.TLSCertToPEM(cert))
		}
	}
}

// NewFromMiniHAR creates a new HAR log from a minihar log.
func NewFromMiniHAR(
	softwareName, softwareVersion string, rs *minihar.RequestSaver,
) *HAR {
	har := &HAR{
		Log: Log{
			Version: "1.2",
			Creator: CreatorInfo{
				Name:    softwareName,
				Version: softwareVersion,
			},
		},
	}
	for _, rts := range rs.RoundTrips {
		entry := new(Entry)
		entry.fillStartedDateTime(rts)
		entry.fillTime(rts)
		entry.fillRequest(rts)
		entry.fillResponse(rts)
		entry.fillTimings(rts)
		entry.fillDNS(rts)
		entry.fillConnect(rts)
		entry.fillTLS(rts)
		har.Log.Entries = append(har.Log.Entries, entry)
	}
	return har
}
