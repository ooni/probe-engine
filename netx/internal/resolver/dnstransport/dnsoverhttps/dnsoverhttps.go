// Package dnsoverhttps implements DNS over HTTPS.
package dnsoverhttps

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
)

// Transport is a DNS over HTTPS modelx.DNSRoundTripper.
//
// As a known bug, this implementation does not cache the domain
// name in the URL for reuse, but this should be easy to fix.
type Transport struct {
	clientDo func(req *http.Request) (*http.Response, error)
	url      string
}

// NewTransport creates a new Transport
func NewTransport(client *http.Client, URL string) *Transport {
	return &Transport{
		clientDo: client.Do,
		url:      URL,
	}
}

// RoundTrip sends a request and receives a response.
func (t *Transport) RoundTrip(ctx context.Context, query []byte) (reply []byte, err error) {
	req, err := http.NewRequest("POST", t.url, bytes.NewReader(query))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/dns-message")
	var resp *http.Response
	resp, err = t.clientDo(req.WithContext(ctx))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// TODO(bassosimone): we should map the status code to a
		// proper Error in the DNS context.
		err = errors.New("doh: server returned error")
		return
	}
	if resp.Header.Get("content-type") != "application/dns-message" {
		err = errors.New("doh: invalid content-type")
		return
	}
	reply, err = ioutil.ReadAll(resp.Body)
	return
}

// RequiresPadding returns true for DoH according to RFC8467
func (t *Transport) RequiresPadding() bool {
	return true
}

// Network returns the transport network (e.g., doh, dot)
func (t *Transport) Network() string {
	return "doh"
}

// Address returns the upstream server address.
func (t *Transport) Address() string {
	return t.url
}
