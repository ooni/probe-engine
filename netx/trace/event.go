package trace

import (
	"crypto/x509"
	"net/http"
	"time"
)

// Event is one of the events within a trace
type Event struct {
	Addresses          []string            `json:",omitempty"`
	Address            string              `json:",omitempty"`
	DNSQuery           []byte              `json:",omitempty"`
	DNSReply           []byte              `json:",omitempty"`
	DataIsTruncated    bool                `json:",omitempty"`
	Data               []byte              `json:",omitempty"`
	Duration           time.Duration       `json:",omitempty"`
	Err                error               `json:",omitempty"`
	HTTPRequest        *http.Request       `json:",omitempty"`
	HTTPResponse       *http.Response      `json:",omitempty"`
	Hostname           string              `json:",omitempty"`
	Name               string              `json:",omitempty"`
	NumBytes           int                 `json:",omitempty"`
	Proto              string              `json:",omitempty"`
	TLSServerName      string              `json:",omitempty"`
	TLSCipherSuite     string              `json:",omitempty"`
	TLSNegotiatedProto string              `json:",omitempty"`
	TLSNextProtos      []string            `json:",omitempty"`
	TLSPeerCerts       []*x509.Certificate `json:",omitempty"`
	TLSVersion         string              `json:",omitempty"`
	Time               time.Time           `json:",omitempty"`
}
