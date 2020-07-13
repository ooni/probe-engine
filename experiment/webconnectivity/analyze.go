package webconnectivity

import (
	"net"
	"net/url"

	"github.com/ooni/probe-engine/netx/modelx"
)

// AnalysisResult contains the results of the analysis performed on the
// client. We obtain it by comparing the measurement and the control.
type AnalysisResult struct {
	DNSConsistency  string  `json:"dns_consistency"`
	BodyLengthMatch *bool   `json:"body_length_match"`
	HeadersMatch    *bool   `json:"header_match"`
	StatusCodeMatch *bool   `json:"status_code_match"`
	TitleMatch      *bool   `json:"title_match"`
	Accessible      *bool   `json:"accessible"`
	Blocking        *string `json:"blocking"`
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
	// 1. flip to consistent if we're targeting an IP address
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
		if key != 0 && (value&inBoth) == inBoth {
			out = consistent
			return
		}
	}
	// 4. when some returned IPs overlap
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
