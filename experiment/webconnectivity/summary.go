package webconnectivity

import (
	"fmt"
	"strings"

	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

// Summary contains the Web Connectivity summary.
type Summary struct {
	Blocking   *string `json:"blocking"`
	Accessible *bool   `json:"accessible"`
}

func stringPointerToString(v *string) (out string) {
	out = "nil"
	if v != nil {
		out = fmt.Sprintf("%+v", *v)
	}
	return
}

// Log logs the summary using the provided logger.
func (s Summary) Log(logger model.Logger) {
	logger.Infof("Blocking %+v", stringPointerToString(s.Blocking))
	logger.Infof("Accessible %+v", boolPointerToString(s.Accessible))
}

// Summarize computes the summary from the TestKeys.
func Summarize(tk *TestKeys) (out Summary) {
	var (
		accessible   = true
		inaccessible = false
		dns          = "dns"
		httpDiff     = "http-diff"
		httpFailure  = "http-failure"
		tcpIP        = "tcp_ip"
	)
	// If the measurement was for an HTTPS website and the HTTP experiment
	// succeded, then either there is a compromised CA in our pool (which is
	// certifi-go), or there is transparent proxying, or we are actually
	// speaking with the legit server. We assume the latter. This applies
	// also to cases in which we are redirected to HTTPS.
	if len(tk.Requests) > 0 && tk.Requests[0].Failure == nil &&
		strings.HasPrefix(tk.Requests[0].Request.URL, "https://") {
		out.Accessible = &accessible
		return
	}
	// If we couldn't contact the control, we cannot do much more here.
	if tk.ControlFailure != nil {
		return
	}
	// If DNS failed with NXDOMAIN and the control is consistent, then it
	// means this website does not exist anymore.
	if tk.DNSExperimentFailure != nil &&
		*tk.DNSExperimentFailure == modelx.FailureDNSNXDOMAINError &&
		tk.DNSConsistency == DNSConsistent {
		return
	}
	// If we tried to connect more than once and never succeded and the
	// DNS is consistent, then it's TCP/IP blocking. Otherwise, it's not
	// unreasonable to assume that the DNS had lied to us.
	if tk.TCPConnectAttempts > 0 && tk.TCPConnectSuccesses <= 0 {
		switch tk.DNSConsistency {
		case DNSConsistent:
			out.Blocking = &tcpIP
			out.Accessible = &inaccessible
		case DNSInconsistent:
			out.Blocking = &dns
			out.Accessible = &inaccessible
		default:
			// this case should not happen with this implementation
			// so it's fine to leave this as unknown
		}
		return
	}
	// If the control failed for HTTP it's not immediate for us to
	// say anything specific on this measurement.
	if tk.Control.HTTPRequest.Failure != nil {
		return
	}
	// Likewise, if we don't have requests to examine, leave it.
	if len(tk.Requests) < 1 {
		return
	}
	// If the HTTP measurement failed there could be a bunch of reasons
	// why this occurred, because of HTTP redirects. Try to guess what
	// could have been wrong by inspecting the error code.
	if tk.Requests[0].Failure != nil {
		switch *tk.Requests[0].Failure {
		case modelx.FailureConnectionRefused:
			// This is possibly because a subsequent connection to some
			// other endpoint has been blocked. So tcp-ip.
			out.Blocking = &tcpIP
			out.Accessible = &inaccessible
		case modelx.FailureConnectionReset:
			// We don't currently support TLS failures and we don't have a
			// way to know if it was during TLS or later. So, for now we are
			// going to call this error condition an http-failure.
			out.Blocking = &httpFailure
			out.Accessible = &inaccessible
		case modelx.FailureDNSNXDOMAINError:
			// This is possibly because a subsequent resolution to
			// some other domain name has been blocked.
			out.Blocking = &dns
			out.Accessible = &inaccessible
		case modelx.FailureEOFError:
			// We have seen this happening with TLS handshakes as well as
			// sometimes with HTTP blocking. So http-failure.
			out.Blocking = &httpFailure
			out.Accessible = &inaccessible
		case modelx.FailureGenericTimeoutError:
			// Alas, here we don't know whether it's connect or whether it's
			// perhaps the TLS handshake. Yet, there is some common ground here
			// that suddenly all our packets are discared at TCP/IP level.
			out.Blocking = &tcpIP
			out.Accessible = &inaccessible
		case modelx.FailureSSLInvalidHostname,
			modelx.FailureSSLInvalidCertificate,
			modelx.FailureSSLUnknownAuthority:
			// We treat these three cases equally. Misconfiguration is a bit
			// less likely since we also checked with the control. Since there
			// is no TLS, for now we're going to call this http-failure.
			out.Blocking = &httpFailure
			out.Accessible = &inaccessible
		default:
			// We have not been able to classify the error. Could this perhaps be
			// caused by a programmer's error? Let us be conservative.
		}
		// So, good that we have classified the error. Yet, how long is the
		// redirect chain? If it's exactly one and we have determined that we
		// should not trust the resolver, then let's bet on the DNS. If the
		// chain is longer, for now better to be conservative. (I would argue
		// that with a lying DNS that's likely the culprit, honestly.)
		if out.Blocking != nil && len(tk.Requests) == 1 &&
			tk.DNSConsistency == DNSInconsistent {
			out.Blocking = &dns
		}
		return
	}
	// So the HTTP request did not fail in the measurement and did not
	// fail in the control as well, didn't it? Then, let us try to guess
	// whether we've got the expected webpage after all. This set of
	// conditions is adapted from MK v0.10.11.
	if tk.StatusCodeMatch != nil && *tk.StatusCodeMatch {
		if tk.BodyLengthMatch != nil && *tk.BodyLengthMatch {
			out.Accessible = &accessible
			return
		}
		if tk.HeadersMatch != nil && *tk.HeadersMatch {
			out.Accessible = &accessible
			return
		}
		if tk.TitleMatch != nil && *tk.TitleMatch {
			out.Accessible = &accessible
			return
		}
	}
	// It seems we didn't get the expected web page. What now? Well, if
	// the DNS does not seem trustworthy, let us blame it.
	if tk.DNSConsistency == DNSInconsistent {
		out.Blocking = &dns
		out.Accessible = &inaccessible
		return
	}
	// The only remaining conclusion seems that the web page we have got
	// doesn't match what we were expecting.
	out.Blocking = &httpDiff
	out.Accessible = &inaccessible
	return
}
