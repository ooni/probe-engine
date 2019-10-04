// Package ootemplate contains OONI templates. That is, data
// structures that are typically included by OONI reports.
package ootemplate

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ooni/netx"
	"github.com/ooni/netx/dnsx"
	"github.com/ooni/netx/handlers"
	"github.com/ooni/netx/model"
)

// QueryAnswer is the answer to a DNS query.
type QueryAnswer struct {
	AnswerType string `json:"answer_type"`
	Hostname   string `json:"hostname,omitempty"`
	IPv4       string `json:"ipv4,omitempty"`
	IPv6       string `json:"ipv6,omitempty"`
}

// QueryEntry contains one of the entries that are part
// of the "queries" key of a OONI report.
type QueryEntry struct {
	Answers          []QueryAnswer `json:"answers"`
	Engine           string        `json:"engine"`
	Failure          *string       `json:"failure"`
	Hostname         string        `json:"hostname"`
	QueryType        string        `json:"query_type"`
	ResolverHostname *string       `json:"resolver_hostname"`
	ResolverPort     *int64        `json:"resolver_port"`
}

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
	Body         HTTPBody            `json:"body"`
	Headers      map[string]string   `json:"headers"`
	Method       string              `json:"method"`
	Tor          HTTPTor             `json:"tor"`
	URL          string              `json:"url"`
	XTrueHeaders map[string][]string `json:"x-true-headers"`
}

// HTTPResponse contains an HTTP response
type HTTPResponse struct {
	Body         HTTPBody            `json:"body"`
	Code         int64               `json:"code"`
	Headers      map[string]string   `json:"headers"`
	XTrueHeaders map[string][]string `json:"x-true-headers"`
}

// RequestsEntry is one of the entries that are part of
// the "requests" key of a OONI report.
type RequestsEntry struct {
	Failure  *string      `json:"failure"`
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
}

// Queries returns the list of events for "queries". The network and
// address arguments are the same of netx.Dialer.NewResolver.
func Queries(
	ctx context.Context, network, address string,
	events [][]model.Measurement,
) []QueryEntry {
	var (
		out      []QueryEntry
		resolver dnsx.Client
	)
	// Attempt at using the same resolver that parent code is
	// using in this experiment. Fallback to system, or die.
	dialer := netx.NewDialer(handlers.NoHandler)
	resolver, err := dialer.NewResolver(network, address)
	if err != nil {
		resolver, err = dialer.NewResolver("system", "")
		if err != nil {
			panic("Cannot create system resolver")
		}
	}
	for _, roundTripEvents := range events {
		for _, ev := range roundTripEvents {
			if ev.Resolve != nil {
				var (
					A    []string
					AAAA []string
				)
				for _, addr := range ev.Resolve.Addresses {
					if strings.Count(addr, ":") > 0 {
						AAAA = append(AAAA, addr)
					} else {
						A = append(A, addr)
					}
				}
				ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
				// Just ignore errors; don't include CNAME on failure
				cname, _ := resolver.LookupCNAME(ctx, ev.Resolve.Hostname)
				out = append(out, QueryEntry{
					Answers:   makeAnswers("A", A, cname),
					Engine:    network,
					Failure:   makeFailure(ev.Resolve.Error),
					Hostname:  ev.Resolve.Hostname,
					QueryType: "A",
				})
				out = append(out, QueryEntry{
					Answers:   makeAnswers("AAAA", AAAA, cname),
					Engine:    network,
					Failure:   makeFailure(ev.Resolve.Error),
					Hostname:  ev.Resolve.Hostname,
					QueryType: "AAAA",
				})
			}
		}
	}
	return out
}

func makeAnswers(
	queryType string, addresses []string, cname string,
) []QueryAnswer {
	var out []QueryAnswer
	for _, addr := range addresses {
		qa := QueryAnswer{AnswerType: queryType}
		if queryType == "A" {
			qa.IPv4 = addr
			out = append(out, qa)
			continue
		}
		if queryType == "AAAA" {
			qa.IPv6 = addr
			out = append(out, qa)
			continue
		}
	}
	if cname != "" {
		out = append(out, QueryAnswer{
			AnswerType: "CNAME",
			Hostname:   cname,
		})
	}
	return out
}

// TCPConnect returns the list of events for "tcp_connect"
func TCPConnect(events [][]model.Measurement) []TCPConnectEntry {
	var out []TCPConnectEntry
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

// Requests returns the list for "requests"
func Requests(events [][]model.Measurement) []RequestsEntry {
	var out []RequestsEntry
	// Note that OONI format wants the most recent request first
	// within the same round-trip, so proceed backwards.
	for idx := len(events) - 1; idx >= 0; idx-- {
		var entry RequestsEntry
		entry.Request.Headers = make(map[string]string)
		entry.Request.XTrueHeaders = make(map[string][]string)
		entry.Response.Headers = make(map[string]string)
		entry.Response.XTrueHeaders = make(map[string][]string)
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
					entry.Request.XTrueHeaders[key] = values
					for _, value := range values {
						entry.Request.Headers[key] = value
						break
					}
				}
				entry.Request.Method = ev.HTTPRequestHeadersDone.Method
				entry.Request.URL = ev.HTTPRequestHeadersDone.URL
				// TODO(bassosimone): do we ever send body?
			}
			if ev.HTTPResponseHeadersDone != nil {
				for key, values := range ev.HTTPResponseHeadersDone.Headers {
					entry.Response.XTrueHeaders[key] = values
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

func makeFailure(err error) (s *string) {
	if err != nil {
		serio := err.Error()
		s = &serio
	}
	return
}
