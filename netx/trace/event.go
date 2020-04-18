package trace

import (
	"crypto/x509"
	"net/http"
	"time"
)

// Snapshot is a snapshot of an HTTP body
type Snapshot struct {
	Data  []byte // actual snapshot
	Limit int64  // max snapshot size
}

// Truncated indicates whether the body is truncated
func (s Snapshot) Truncated() bool {
	return int64(len(s.Data)) >= s.Limit
}

func (s Snapshot) String() string {
	return string(s.Data)
}

// Event is one of the events within a trace
type Event struct {
	Addresses          []string            `json:",omitempty"`
	Address            string              `json:",omitempty"`
	DNSQuery           []byte              `json:",omitempty"`
	DNSReply           []byte              `json:",omitempty"`
	Data               []byte              `json:",omitempty"`
	Duration           time.Duration       `json:",omitempty"`
	Err                error               `json:",omitempty"`
	HTTPRequestBody    *Snapshot           `json:",omitempty"`
	HTTPRequest        *http.Request       `json:",omitempty"`
	HTTPResponseBody   *Snapshot           `json:",omitempty"`
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
