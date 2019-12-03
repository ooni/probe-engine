// Package oodatamodel contains the OONI data model.
package oodatamodel

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"strconv"
	"unicode/utf8"

	"github.com/ooni/netx/x/porcelain"
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
func NewTCPConnectList(results porcelain.Results) TCPConnectList {
	var out TCPConnectList
	for _, connect := range results.Connects {
		// We assume Go is passing us legit data structs
		ip, sport, _ := net.SplitHostPort(connect.RemoteAddress)
		iport, _ := strconv.Atoi(sport)
		out = append(out, TCPConnectEntry{
			IP:   ip,
			Port: iport,
			Status: TCPConnectStatus{
				Failure: makeFailure(connect.Error),
				Success: connect.Error == nil,
			},
		})
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

// MaybeBinaryValue is a possibly binary string. We use this helper class
// to define a custom JSON encoder that allows us to choose the proper
// representation depending on whether the Value field is valid UTF-8 or not.
type MaybeBinaryValue struct {
	Value string
}

// MarshalJSON marshal the body to JSON following the OONI spec that says
// that UTF-8 bodies are represened as string and non-UTF-8 bodies are
// instead represented as `{"format":"base64","data":"..."}`.
func (hb MaybeBinaryValue) MarshalJSON() ([]byte, error) {
	if utf8.ValidString(hb.Value) {
		return json.Marshal(hb.Value)
	}
	er := make(map[string]string)
	er["format"] = "base64"
	er["data"] = base64.StdEncoding.EncodeToString([]byte(hb.Value))
	return json.Marshal(er)
}

// HTTPBody is an HTTP body.
type HTTPBody MaybeBinaryValue

// HTTPHeaders contains HTTP headers.
type HTTPHeaders map[string]MaybeBinaryValue

// HTTPRequest contains an HTTP request
type HTTPRequest struct {
	Body            HTTPBody    `json:"body"`
	BodyIsTruncated bool        `json:"body_is_truncated"`
	Headers         HTTPHeaders `json:"headers"`
	Method          string      `json:"method"`
	Tor             HTTPTor     `json:"tor"`
	URL             string      `json:"url"`
}

// HTTPResponse contains an HTTP response
type HTTPResponse struct {
	Body            HTTPBody    `json:"body"`
	BodyIsTruncated bool        `json:"body_is_truncated"`
	Code            int64       `json:"code"`
	Headers         HTTPHeaders `json:"headers"`
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
func NewRequestList(httpresults *porcelain.HTTPDoResults) RequestList {
	// TODO(bassosimone): here I'm using netx snapshots which are
	// limited to 1<<20. They are probably good enough for a really
	// wide range of cases, and truncating the body seems good for
	// loading measurements on mobile as well. I think I should make
	// sure I modify the documentation to mention that.
	var out RequestList
	if httpresults == nil {
		return out
	}
	in := httpresults.TestKeys.HTTPRequests
	// OONI's data format wants more recent request first
	for idx := len(in) - 1; idx >= 0; idx-- {
		var entry RequestEntry
		entry.Failure = makeFailure(in[idx].Error)
		entry.Request.Headers = make(HTTPHeaders)
		for key, values := range in[idx].RequestHeaders {
			for _, value := range values {
				entry.Request.Headers[key] = MaybeBinaryValue{
					Value: value,
				}
				// We skip processing after the first header with
				// such name has been processed. This is a known
				// issue of OONI's data model.
				break
			}
		}
		entry.Request.Method = in[idx].RequestMethod
		entry.Request.URL = in[idx].RequestURL
		entry.Request.Body.Value = string(in[idx].RequestBodySnap)
		entry.Request.BodyIsTruncated = in[idx].MaxBodySnapSize > 0 &&
			int64(len(in[idx].RequestBodySnap)) >= in[idx].MaxBodySnapSize
		entry.Response.Headers = make(HTTPHeaders)
		for key, values := range in[idx].ResponseHeaders {
			for _, value := range values {
				entry.Response.Headers[key] = MaybeBinaryValue{
					Value: value,
				}
				// We skip processing after the first header with
				// such name has been processed. This is a known
				// issue of OONI's data model.
				break
			}
		}
		entry.Response.Code = in[idx].ResponseStatusCode
		entry.Response.Body.Value = string(in[idx].ResponseBodySnap)
		entry.Response.BodyIsTruncated = in[idx].MaxBodySnapSize > 0 &&
			int64(len(in[idx].ResponseBodySnap)) >= in[idx].MaxBodySnapSize
		out = append(out, entry)
	}
	return out
}
