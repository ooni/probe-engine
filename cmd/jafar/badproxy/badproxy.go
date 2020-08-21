// Package badproxy contains a bad proxy. Specifically this proxy
// will read some bytes from the input and then close the connection.
package badproxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/google/martian/v3/mitm"
)

// CensoringProxy is a bad proxy
type CensoringProxy struct {
	mitmNewAuthority func(
		name string, organization string,
		validity time.Duration,
	) (*x509.Certificate, *rsa.PrivateKey, error)

	mitmNewConfig func(
		ca *x509.Certificate, privateKey interface{},
	) (*mitm.Config, error)

	tlsListen func(
		network string, laddr string, config *tls.Config,
	) (net.Listener, error)
}

// NewCensoringProxy creates a new bad proxy
func NewCensoringProxy() *CensoringProxy {
	return &CensoringProxy{
		mitmNewAuthority: mitm.NewAuthority,
		mitmNewConfig:    mitm.NewConfig,
		tlsListen:        tls.Listen,
	}
}

func (p *CensoringProxy) serve(conn net.Conn) {
	deadline := time.Now().Add(250 * time.Millisecond)
	conn.SetDeadline(deadline)
	// To simulate the case where the proxy isn't willing to forward our
	// traffic, we close the connection (1) right after the handshake for
	// TLS connections and (2) reasonably after we've received the HTTP
	// request for cleartext connections. This may break in several cases
	// but is good enough approximation of these bad proxies for now.
	if tlsconn, ok := conn.(*tls.Conn); ok {
		tlsconn.Handshake()
	} else {
		const maxread = 1 << 17
		reader := io.LimitReader(conn, maxread)
		ioutil.ReadAll(reader)
	}
	conn.Close()
}

func (p *CensoringProxy) run(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil && strings.Contains(
			err.Error(), "use of closed network connection") {
			return
		}
		if err == nil {
			// It's difficult to make accept fail, so restructure
			// the code such that we enter into the happy path
			go p.serve(conn)
		}
	}
}

// Start starts the bad proxy for TCP.
func (p *CensoringProxy) Start(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	go p.run(listener)
	return listener, nil
}

// StartTLS starts the bad proxy for TLS.
func (p *CensoringProxy) StartTLS(address string) (net.Listener, *x509.Certificate, error) {
	cert, privkey, err := p.mitmNewAuthority(
		"jafar", "OONI", 24*time.Hour,
	)
	if err != nil {
		return nil, nil, err
	}
	config, err := p.mitmNewConfig(cert, privkey)
	if err != nil {
		return nil, nil, err
	}
	listener, err := p.tlsListen("tcp", address, config.TLS())
	if err != nil {
		return nil, nil, err
	}
	go p.run(listener)
	return listener, cert, nil
}
