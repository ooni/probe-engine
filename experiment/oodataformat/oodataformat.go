// Package oodataformat contains the OONI data format.
package oodataformat

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"unicode/utf8"

	"github.com/ooni/netx/model"
)

// TCPConnectStatus contains the TCP connect status.
type TCPConnectStatus struct {
	Failure *string `json:"failure"`
	Success bool    `json:"success"`
}

// TCPConnectEntry contains one of the entries that are part
// of the "tcp_connect" key of a OONI report.
type TCPConnectEntry struct {
	IP     string           `json:"ip"`
	Port   int              `json:"port"`
	Status TCPConnectStatus `json:"status"`
}

// TCPConnectList is a list of TCPConnectEntry
type TCPConnectList []TCPConnectEntry

// NewTCPConnectList creates a new TCPConnectList
func NewTCPConnectList(events [][]model.Measurement) TCPConnectList {
	var out TCPConnectList
	for _, roundTripEvents := range events {
		for _, ev := range roundTripEvents {
			if ev.Connect != nil {
				// We assume Go is passing us legit data structs
				ip, sport, err := net.SplitHostPort(ev.Connect.RemoteAddress)
				if err != nil {
					continue
				}
				iport, err := strconv.Atoi(sport)
				if err != nil {
					continue
				}
				if iport < 0 || iport > 65535 {
					continue
				}
				out = append(out, TCPConnectEntry{
					IP:   ip,
					Port: iport,
					Status: TCPConnectStatus{
						Failure: makeFailure(ev.Connect.Error),
						Success: ev.Connect.Error == nil,
					},
				})
			}
		}
	}
	return out
}

func makeFailure(err error) (s *string) {
	if err != nil {
		serio := err.Error()
		s = &serio
	}
	return
}

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
	Body    HTTPBody          `json:"body"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
	Tor     HTTPTor           `json:"tor"`
	URL     string            `json:"url"`
}

// HTTPResponse contains an HTTP response
type HTTPResponse struct {
	Body    HTTPBody          `json:"body"`
	Code    int64             `json:"code"`
	Headers map[string]string `json:"headers"`
}

// RequestEntry is one of the entries that are part of
// the "requests" key of a OONI report.
type RequestEntry struct {
	Failure  *string      `json:"failure"`
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}

// RequestList is a list of RequestEntry
type RequestList []RequestEntry

// NewRequestList returns the list for "requests"
func NewRequestList(events [][]model.Measurement) RequestList {
	var out RequestList
	// within the same round-trip, so proceed backwards.
	for idx := len(events) - 1; idx >= 0; idx-- {
		var entry RequestEntry
		entry.Request.Headers = make(map[string]string)
		entry.Response.Headers = make(map[string]string)
		for _, ev := range events[idx] {
			// Note how dividing events by round trip simplifies
			// deciding whether there has been an error
			if ev.Resolve != nil && ev.Resolve.Error != nil {
				entry.Failure = makeFailure(ev.Resolve.Error)
			}
			if ev.Connect != nil && ev.Connect.Error != nil {
				entry.Failure = makeFailure(ev.Connect.Error)
			}
			if ev.Read != nil && ev.Read.Error != nil {
				entry.Failure = makeFailure(ev.Read.Error)
			}
			if ev.Write != nil && ev.Write.Error != nil {
				entry.Failure = makeFailure(ev.Write.Error)
			}
			if ev.HTTPRequestHeadersDone != nil {
				for key, values := range ev.HTTPRequestHeadersDone.Headers {
					for _, value := range values {
						entry.Request.Headers[key] = value
						break
					}
				}
				entry.Request.Method = ev.HTTPRequestHeadersDone.Method
				entry.Request.URL = ev.HTTPRequestHeadersDone.URL
				// TODO(bassosimone): do we ever send body? We should
				// probably have an issue for this after merging.
			}
			if ev.HTTPResponseHeadersDone != nil {
				for key, values := range ev.HTTPResponseHeadersDone.Headers {
					for _, value := range values {
						entry.Response.Headers[key] = value
						break
					}
				}
				entry.Response.Code = ev.HTTPResponseHeadersDone.StatusCode
			}
			if ev.HTTPResponseBodyPart != nil {
				// Note that it's legal Go code to return bytes _and_ an
				// error, e.g. EOF, from an io.Reader. So, we need to process
				// the bytes anyway and then we can check the error.
				entry.Response.Body.Value += string(ev.HTTPResponseBodyPart.Data)
				if ev.HTTPResponseBodyPart.Error != nil &&
					ev.HTTPResponseBodyPart.Error != io.EOF {
					// We may see error here if we receive a bad TLS record or
					// bad gzip data. ReadEvent only sees what happens in the
					// network. Here we sit on top of much more stuff.
					entry.Failure = makeFailure(ev.HTTPResponseBodyPart.Error)
				}
			}
		}
		out = append(out, entry)
	}
	return out
}
