package resolver

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/netx/modelx"
)

// DNSOverUDP is a DNS over UDP RoundTripper.
type DNSOverUDP struct {
	dialer  modelx.Dialer
	address string
}

// NewDNSOverUDP creates a new Transport
func NewDNSOverUDP(dialer modelx.Dialer, address string) *DNSOverUDP {
	return &DNSOverUDP{
		dialer:  dialer,
		address: address,
	}
}

// RoundTrip sends a request and receives a response.
func (t *DNSOverUDP) RoundTrip(ctx context.Context, query []byte) (reply []byte, err error) {
	conn, err := t.dialer.DialContext(ctx, "udp", t.address)
	if err != nil {
		return
	}
	defer conn.Close()
	// Use five seconds timeout like Bionic does. See
	// https://labs.ripe.net/Members/baptiste_jonglez_1/persistent-dns-connections-for-reliability-and-performance
	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return
	}
	_, err = conn.Write(query)
	if err != nil {
		return
	}
	reply = make([]byte, 1<<17)
	var n int
	n, err = conn.Read(reply)
	if err == nil {
		reply = reply[:n]
	}
	return
}

// RequiresPadding returns false for UDP according to RFC8467
func (t *DNSOverUDP) RequiresPadding() bool {
	return false
}

// Network returns the transport network (e.g., doh, dot)
func (t *DNSOverUDP) Network() string {
	return "udp"
}

// Address returns the upstream server address.
func (t *DNSOverUDP) Address() string {
	return t.address
}
