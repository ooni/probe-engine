// Package oodatamodel contains the OONI data model.
package oodatamodel

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
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

// MarshalJSON marshals the body to JSON following the OONI spec that says
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

// HTTPBody is an HTTP body. As an implementation note, this type must be
// an alias for the MaybeBinaryValue type, otherwise the specific serialisation
// mechanism implemented by MaybeBinaryValue is not working.
type HTTPBody = MaybeBinaryValue

// HTTPHeaders contains HTTP headers. This headers representation is
// deprecated in favour of HTTPHeadersList since data format 0.3.0.
type HTTPHeaders map[string]MaybeBinaryValue

// HTTPHeader is a single HTTP header.
type HTTPHeader struct {
	Key   string
	Value MaybeBinaryValue
}

// MarshalJSON marshals the body to JSON following the OONI spec that says
// that UTF-8 bodies are represened as string and non-UTF-8 bodies are
// instead represented as `{"format":"base64","data":"..."}`.
func (hh HTTPHeader) MarshalJSON() ([]byte, error) {
	if utf8.ValidString(hh.Value.Value) {
		return json.Marshal([]string{hh.Key, hh.Value.Value})
	}
	value := make(map[string]string)
	value["format"] = "base64"
	value["data"] = base64.StdEncoding.EncodeToString([]byte(hh.Value.Value))
	return json.Marshal([]interface{}{hh.Key, value})
}

// HTTPHeadersList is a list of headers.
type HTTPHeadersList []HTTPHeader

// HTTPRequest contains an HTTP request.
//
// Headers are a map in Web Connectivity data format but
// we have added support for a list since data format version
// equal to 0.2.1 (later renamed to 0.3.0).
type HTTPRequest struct {
	Body            HTTPBody        `json:"body"`
	BodyIsTruncated bool            `json:"body_is_truncated"`
	HeadersList     HTTPHeadersList `json:"headers_list"`
	Headers         HTTPHeaders     `json:"headers"`
	Method          string          `json:"method"`
	Tor             HTTPTor         `json:"tor"`
	URL             string          `json:"url"`
}

// HTTPResponse contains an HTTP response.
//
// Headers are a map in Web Connectivity data format but
// we have added support for a list since data format version
// equal to 0.2.1 (later renamed to 0.3.0).
type HTTPResponse struct {
	Body            HTTPBody        `json:"body"`
	BodyIsTruncated bool            `json:"body_is_truncated"`
	Code            int64           `json:"code"`
	HeadersList     HTTPHeadersList `json:"headers_list"`
	Headers         HTTPHeaders     `json:"headers"`
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

func addheaders(
	source http.Header,
	destList *HTTPHeadersList,
	destMap *HTTPHeaders,
) {
	for key, values := range source {
		for index, value := range values {
			value := MaybeBinaryValue{Value: value}
			if index == 0 {
				(*destMap)[key] = value
			}
			*destList = append(*destList, HTTPHeader{
				Key:   key,
				Value: value,
			})
		}
	}
}

// NewRequestList returns the list for "requests"
func NewRequestList(httpresults *porcelain.HTTPDoResults) RequestList {
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
		addheaders(
			in[idx].RequestHeaders, &entry.Request.HeadersList,
			&entry.Request.Headers,
		)
		entry.Request.Method = in[idx].RequestMethod
		entry.Request.URL = in[idx].RequestURL
		entry.Request.Body.Value = string(in[idx].RequestBodySnap)
		entry.Request.BodyIsTruncated = in[idx].MaxBodySnapSize > 0 &&
			int64(len(in[idx].RequestBodySnap)) >= in[idx].MaxBodySnapSize
		entry.Response.Headers = make(HTTPHeaders)
		addheaders(
			in[idx].ResponseHeaders, &entry.Response.HeadersList,
			&entry.Response.Headers,
		)
		entry.Response.Code = in[idx].ResponseStatusCode
		entry.Response.Body.Value = string(in[idx].ResponseBodySnap)
		entry.Response.BodyIsTruncated = in[idx].MaxBodySnapSize > 0 &&
			int64(len(in[idx].ResponseBodySnap)) >= in[idx].MaxBodySnapSize
		out = append(out, entry)
	}
	return out
}
