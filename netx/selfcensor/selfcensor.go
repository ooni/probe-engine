// Package selfcensor contains code that triggers censorship. We use
// this functionality to implement integration tests.
//
// Jafar is the evil grand vizier of censorship. It has been implemented
// at github.com/ooni/jafar. Libjafar brings some jafar concepts inside
// the github.com/ooni/probe-engine repository. While the code in the jafar
// repository requires Linux for most cool stuff, this code is portable
// to all systems. Yet, the code in here works by changing the probe-engine
// internals, so it is definitely a much less accurate mechanism.
//
// To use this library, you need to use `-tags selfcensor`. If this that is
// specified, them the code will honour the MINIOONI_SELFCENSOR_SPEC environment
// variable and will implement the censorship policy described by its
// content. Such variable shall contain a JSON serialized Spec structure.
//
// Examples
//
// The following example causes NXDOMAIN to be returned for `dns.google`:
//
//     export MINIOONI_SELFCENSOR_SPEC='{"PoisonSystemDNS":{"dns.google":["NXDOMAIN"]}}'
//
// The following example blocks connecting to `8.8.8.8:443`:
//
//     export MINIOONI_SELFCENSOR_SPEC='{"BlockedEndpoints":{"8.8.8.8:443":"REJECT"}}'
//
// The following example blocks packets containing dns.google:
//
//     export MINIOONI_SELFCENSOR_SPEC='{"BlockedFingerprints":{"dns.google":"RST"}}'
//
// The documentation of the Spec structure contains further information.
package selfcensor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/ooni/probe-engine/atomicx"
	"github.com/ooni/probe-engine/internal/runtimex"
)

// Spec contains the spec for minijafar. This spec should be passed as a
// serialized JSON in the MINIOONI_SELFCENSOR_SPEC environment variable.
type Spec struct {
	// PoisonSystemDNS allows you to change the behaviour of the system
	// DNS regarding specific domains. They keys are the domains and the
	// values are the IP addresses to return. If you set the values for
	// a domain to `[]string{"NXDOMAIN"}`, the system resolver will return
	// an NXDOMAIN response. If you set the values for a domain to
	// `[]string{"TIMEOUT"}` the system resolver will block until the
	// context used by the code is expired.
	PoisonSystemDNS map[string][]string

	// BlockedEndpoints allows you to block specific IP endpoints. The key is
	// `IP:port` to block. The format is the same of net.JoinHostPort. If
	// the value is "REJECT", then the connection attempt will fail with
	// ECONNREFUSED. If the value is "TIMEOUT", then the connector will block
	// until the context is expired. If the value is anything else, we
	// will perform a "REJECT".
	BlockedEndpoints map[string]string

	// BlockedFingerprints allows you to block packets whose body contains
	// specific fingerprints. Of course, the key is the fingerprint. If
	// the value is "RST", then the connection will be reset. If the value
	// is "TIMEOUT", then the code will block until the context is
	// expired. If the value is anything else, we will perform a "RST".
	BlockedFingerprints map[string]string
}

// EnvironmentVariable is the name of the environment variable that you
// must set in order to enable selfcensor functionality.
const EnvironmentVariable = "MINIOONI_SELFCENSOR_SPEC"

var (
	enabled *atomicx.Int64
	spec    *Spec
	mu      sync.Mutex
)

func init() {
	enabled = atomicx.NewInt64()
	env := getenv(EnvironmentVariable)
	if env == "" {
		return // no environment variable or no `-tags selfcensor`
	}
	spec = new(Spec)
	runtimex.PanicOnError(
		json.Unmarshal([]byte(env), spec), "selfcensor: cannot parse spec")
	enabled.Add(1)
	log.Printf("selfcensor: enabled and using this spec: %s", env)
}

// SystemResolver is selfcensor system resolver. If MINIOONI_SELFCENSOR_SPEC is
// set, it will use its content to censor the system resolver.
type SystemResolver struct{}

// LookupHost implements Resolver.LookupHost
func (r SystemResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	if enabled.Load() != 0 { // jumps not taken by default
		mu.Lock()
		defer mu.Unlock()
		if spec.PoisonSystemDNS != nil {
			values := spec.PoisonSystemDNS[hostname]
			if len(values) == 1 && values[0] == "NXDOMAIN" {
				return nil, errors.New("no such host")
			}
			if len(values) == 1 && values[0] == "TIMEOUT" {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			if len(values) > 0 {
				return values, nil
			}
		}
		// FALLTHROUGH
	}
	return net.DefaultResolver.LookupHost(ctx, hostname)
}

// Network implements Resolver.Network
func (r SystemResolver) Network() string {
	return "system"
}

// Address implements Resolver.Address
func (r SystemResolver) Address() string {
	return ""
}

// SystemDialer is selfcensor system dialer. If MINIOONI_SELFCENSOR_SPEC is
// set, it will use its content to censor the system dialer.
type SystemDialer struct{}

// defaultDialer is the dialer we use by default
var defaultDialer = new(net.Dialer)

// DialContext implemnts Dialer.DialContext
func (d SystemDialer) DialContext(
	ctx context.Context, network, address string) (net.Conn, error) {
	if enabled.Load() != 0 { // jumps not taken by default
		mu.Lock()
		defer mu.Unlock()
		if spec.BlockedEndpoints != nil {
			action, ok := spec.BlockedEndpoints[address]
			if ok && action == "TIMEOUT" {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			if ok {
				switch network {
				case "tcp", "tcp4", "tcp6":
					return nil, errors.New("connection refused")
				default:
					// not applicable
				}
			}
		}
		if spec.BlockedFingerprints != nil {
			conn, err := defaultDialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			return connWrapper{Conn: conn, closed: make(chan interface{}),
				fingerprints: spec.BlockedFingerprints}, nil
		}
		// FALLTHROUGH
	}
	return defaultDialer.DialContext(ctx, network, address)
}

type connWrapper struct {
	net.Conn
	closed       chan interface{}
	fingerprints map[string]string
}

func (c connWrapper) Read(p []byte) (int, error) {
	// TODO(bassosimone): implement reassembly to workaround the
	// splitting of the ClientHello message.
	count, err := c.Conn.Read(p)
	if err != nil {
		return 0, err
	}
	return c.match(p, count)
}

func (c connWrapper) Write(p []byte) (int, error) {
	if _, err := c.match(p, len(p)); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}

func (c connWrapper) match(p []byte, n int) (int, error) {
	p = p[:n] // trim
	for key, value := range c.fingerprints {
		if bytes.Index(p, []byte(key)) != -1 {
			if value == "TIMEOUT" {
				<-c.closed
				return 0, errors.New("use of closed network connection")
			}
			return 0, errors.New("connection reset by peer")
		}
	}
	return n, nil
}

func (c connWrapper) Close() error {
	c.closed <- true
	return c.Conn.Close()
}
