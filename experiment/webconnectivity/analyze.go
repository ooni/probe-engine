package webconnectivity

import (
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ooni/probe-engine/netx/modelx"
)

// AnalysisResult contains the results of the analysis performed on the
// client. We obtain it by comparing the measurement and the control.
type AnalysisResult struct {
	DNSConsistency   string   `json:"dns_consistency"`
	BlockedEndpoints []string `json:"x_blocked_endpoints"` // not in spec
	BodyLengthMatch  *bool    `json:"body_length_match"`
	BodyProportion   *float64 `json:"body_proportion"`
	HeadersMatch     *bool    `json:"header_match"`
	StatusCodeMatch  *bool    `json:"status_code_match"`
	TitleMatch       *bool    `json:"title_match"`
	Accessible       *bool    `json:"accessible"`
	Blocking         *string  `json:"blocking"`
}

// Analyze performs follow-up analysis on the webconnectivity measurement by
// comparing the measurement (tk.TestKeys) and the control (tk.Control). This
// function will return the results of the analysis.
func Analyze(target string, tk *TestKeys) (out AnalysisResult) {
	URL, err := url.Parse(target)
	if err != nil {
		return // should not happen in practice
	}
	out.DNSConsistency = DNSConsistency(URL, tk)
	out.BlockedEndpoints = BlockedEndpoints(tk)
	out.BodyLengthMatch, out.BodyProportion = BodyLengthChecks(tk)
	out.StatusCodeMatch = StatusCodeMatch(tk)
	out.HeadersMatch = HeadersMatch(tk)
	out.TitleMatch = TitleMatch(tk)
	return
}

// ControlDNSNameError is the error returned by the control on NXDOMAIN
const ControlDNSNameError = "dns_name_error"

// DNSConsistency returns the value of the AnalysisResult.DNSConsistency
// field that you should set into an AnalysisResult struct.
//
// This implementation is a simplified version of the implementation of
// the same check in Measurement Kit v0.10.11.
func DNSConsistency(URL *url.URL, tk *TestKeys) (out string) {
	// 0. start assuming it's not consistent
	const (
		consistent   = "consistent"
		inconsistent = "inconsistent"
	)
	out = inconsistent
	// 1. flip to consistent if we're targeting an IP address because the
	// control will actually return dns_name_error in this case.
	if net.ParseIP(URL.Hostname()) != nil {
		out = consistent
		return
	}
	// 2. flip to consistent if the failures are compatible
	if tk.DNSExperimentFailure != nil && tk.Control.DNS.Failure != nil {
		switch *tk.Control.DNS.Failure {
		case ControlDNSNameError: // the control returns this on NXDOMAIN error
			switch *tk.DNSExperimentFailure {
			case modelx.FailureDNSNXDOMAINError:
				out = consistent
			}
		}
		return
	}
	// 3. flip to consistent if measurement and control returned IP addresses
	// that belong to the same Autonomous System(s).
	//
	// This specific check is present in MK's implementation.
	//
	// Note that this covers also the cases where the measurement contains only
	// bogons while the control does not contain bogons.
	//
	// Note that this also covers the cases where results are equal.
	const (
		inMeasurement = 1 << 0
		inControl     = 1 << 1
		inBoth        = inMeasurement | inControl
	)
	asnmap := make(map[int64]int)
	for _, entry := range tk.Queries {
		for _, answer := range entry.Answers {
			asnmap[answer.ASN] |= inMeasurement
		}
	}
	for _, asn := range tk.Control.DNS.ASNs {
		asnmap[asn] |= inControl
	}
	for key, value := range asnmap {
		// zero means that ASN lookup failed
		if key != 0 && (value&inBoth) == inBoth {
			out = consistent
			return
		}
	}
	// 4. when ASN lookup failed (unlikely), check whether
	// there is overlap in the returned IP addresses
	ipmap := make(map[string]int)
	for _, entry := range tk.Queries {
		for _, answer := range entry.Answers {
			// we exclude the case of empty string below
			ipmap[answer.IPv4] |= inMeasurement
			ipmap[answer.IPv6] |= inMeasurement
		}
	}
	for _, ip := range tk.Control.DNS.Addrs {
		ipmap[ip] |= inControl
	}
	for key, value := range ipmap {
		if key != "" && (value&inBoth) == inBoth {
			out = consistent
			return
		}
	}
	// 5. conclude that measurement and control are inconsistent
	return
}

// BlockedEndpoints computes which endpoints are blocked by comparing
// what the measurement and control found to be blocked.
//
// This is not done by the original implementations. They used to
// record this information inside of the `tcp_connect` result in the
// measurement as `blocked`. This implementation instead writes a
// list of blocked TCP endpoints, because:
//
// 1. it is dirty to stuff the result of analysis inside of the
// measurement and we agreed multiple times that we were going to
// avoid mixing measurement and analysis
//
// 2. it is more practical to parse a toplevel array both when
// parsing through a script and when doing it manually
//
// 3. it is complex to implement the original behavior in Go
func BlockedEndpoints(tk *TestKeys) []string {
	out := []string{}
	for _, measurement := range tk.TCPConnect {
		epnt := net.JoinHostPort(measurement.IP, strconv.Itoa(measurement.Port))
		if control, ok := tk.Control.TCPConnect[epnt]; ok {
			if control.Failure == nil && measurement.Status.Failure != nil {
				out = append(out, epnt)
			}
		}
	}
	return out
}

// BodyLengthChecks returns whether the measured body is reasonably
// long as much as the control body as well as the proportion between
// the two bodies. This check may return nil, nil when such a
// comparison would actually not be applicable.
func BodyLengthChecks(tk *TestKeys) (match *bool, percentage *float64) {
	control := tk.Control.HTTPRequest.BodyLength
	if control <= 0 {
		return
	}
	if len(tk.Requests) <= 0 {
		return
	}
	response := tk.Requests[0].Response
	if response.BodyIsTruncated {
		return
	}
	measurement := int64(len(response.Body.Value))
	if measurement <= 0 {
		return
	}
	const bodyProportionFactor = 0.7
	var proportion float64
	if measurement >= control {
		proportion = float64(control) / float64(measurement)
	} else {
		proportion = float64(measurement) / float64(control)
	}
	v := proportion > bodyProportionFactor
	match = &v
	percentage = &proportion
	return
}

// StatusCodeMatch returns whether the status code of the measurement
// matches the status code of the control, or nil if such comparison
// is actually not applicable.
func StatusCodeMatch(tk *TestKeys) (out *bool) {
	control := tk.Control.HTTPRequest.StatusCode
	measurement := tk.HTTPResponseStatus
	if control == 0 || measurement == 0 {
		return // no real status code
	}
	value := control == measurement
	if value == true {
		// if the status codes are equal, they clearly match
		out = &value
		return
	}
	// This fix is part of Web Connectivity in MK and in Python since
	// basically forever; my recollection is that we want to work around
	// cases where the test helper is failing(?!). Unlike previous
	// implementations, this implementation avoids a false positive
	// when both measurement and control statuses are 500.
	if control/100 == 5 {
		return
	}
	out = &value
	return
}

// HeadersMatch returns whether uncommon headers match between control and
// measurement, or nil if check is not applicable.
func HeadersMatch(tk *TestKeys) *bool {
	if len(tk.Requests) <= 0 {
		return nil
	}
	if tk.Requests[0].Response.Code == 0 {
		return nil
	}
	if tk.Control.HTTPRequest.StatusCode == 0 {
		return nil
	}
	control := tk.Control.HTTPRequest.Headers
	// Implementation note: using map because we only care about the
	// keys being different and we ignore the values.
	measurement := tk.Requests[0].Response.Headers
	// Rather than checking all headers first and then uncommon headers
	// just check whether the uncommon headers are matching
	const (
		inMeasurement = 1 << 0
		inControl     = 1 << 1
		inBoth        = inMeasurement | inControl
	)
	commonHeaders := map[string]bool{
		"date":                      true,
		"content-type":              true,
		"server":                    true,
		"cache-control":             true,
		"vary":                      true,
		"set-cookie":                true,
		"location":                  true,
		"expires":                   true,
		"x-powered-by":              true,
		"content-encoding":          true,
		"last-modified":             true,
		"accept-ranges":             true,
		"pragma":                    true,
		"x-frame-options":           true,
		"etag":                      true,
		"x-content-type-options":    true,
		"age":                       true,
		"via":                       true,
		"p3p":                       true,
		"x-xss-protection":          true,
		"content-language":          true,
		"cf-ray":                    true,
		"strict-transport-security": true,
		"link":                      true,
		"x-varnish":                 true,
	}
	matching := make(map[string]int)
	for key := range measurement {
		if _, ok := commonHeaders[key]; !ok {
			matching[strings.ToLower(key)] |= inMeasurement
		}
	}
	for key := range control {
		if _, ok := commonHeaders[key]; !ok {
			matching[strings.ToLower(key)] |= inControl
		}
	}
	good := true
	for _, value := range matching {
		if (value & inBoth) != inBoth {
			good = false
			break
		}
	}
	return &good
}

// TitleMatch returns whether the measurement and the control titles
// reasonably match, or nil if not applicable.
func TitleMatch(tk *TestKeys) (out *bool) {
	if len(tk.Requests) <= 0 {
		return
	}
	response := tk.Requests[0].Response
	if response.Code == 0 {
		return
	}
	if response.BodyIsTruncated {
		return
	}
	if tk.Control.HTTPRequest.StatusCode == 0 {
		return
	}
	control := tk.Control.HTTPRequest.Title
	measurementBody := response.Body.Value
	re := regexp.MustCompile(`(?i)<title>([^<]{1,128})</title>`) // like MK
	v := re.FindStringSubmatch(measurementBody)
	if len(v) < 2 {
		return
	}
	measurement := v[1]
	const (
		inMeasurement = 1 << 0
		inControl     = 1 << 1
		inBoth        = inMeasurement | inControl
	)
	words := make(map[string]int)
	// We don't consider to match words that are shorter than 5
	// characters (5 is the average word length for english)
	//
	// The original implementation considered the word order but
	// considering different languages it seems we could have less
	// false positives by ignoring the word order.
	const minWordLength = 5
	for _, word := range strings.Split(measurement, " ") {
		if len(word) >= minWordLength {
			words[strings.ToLower(word)] |= inMeasurement
		}
	}
	for _, word := range strings.Split(control, " ") {
		if len(word) >= minWordLength {
			words[strings.ToLower(word)] |= inControl
		}
	}
	good := true
	for _, score := range words {
		if (score & inBoth) != inBoth {
			good = false
			break
		}
	}
	return &good
}
