package webconnectivity

import (
	"net"
	"net/url"

	"github.com/ooni/probe-engine/internal/runtimex"
)

// Endpoint describes a TCP/TLS endpoint.
type Endpoint struct {
	String       string // String representation
	URLGetterURL string // URL for urlgetter
}

// NewEndpoints creates a list of TCP/TLS endpoints to test from the
// target URL and the list of resolved IP addresses.
func NewEndpoints(URL *url.URL, addrs []string) (out []Endpoint) {
	out = []Endpoint{}
	port := NewEndpointPort(URL)
	for _, addr := range addrs {
		endpoint := net.JoinHostPort(addr, port.Port)
		out = append(out, Endpoint{
			String:       endpoint,
			URLGetterURL: (&url.URL{Scheme: port.URLGetterScheme, Host: endpoint}).String(),
		})
	}
	return
}

// EndpointPort is the port to be used by a TCP/TLS endpoint.
type EndpointPort struct {
	URLGetterScheme string
	Port            string
}

// NewEndpointPort creates an EndpointPort from the given URL. This function
// panic if the scheme is not `http` or `https` as well as if the host is not
// valid. The latter should not happen if you used url.Parse.
func NewEndpointPort(URL *url.URL) (out EndpointPort) {
	if URL.Scheme != "http" && URL.Scheme != "https" {
		panic("passed an unexpected scheme")
	}
	switch URL.Scheme {
	case "http":
		out.URLGetterScheme, out.Port = "tcpconnect", "80"
	case "https":
		out.URLGetterScheme, out.Port = "tlshandshake", "443"
	}
	if URL.Host != URL.Hostname() {
		_, port, err := net.SplitHostPort(URL.Host)
		runtimex.PanicOnError(err, "SplitHostPort should not fail here")
		out.Port = port
	}
	return
}
