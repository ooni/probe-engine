package enginelocate

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/pion/stun"
)

func TestSTUNIPLookupCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // stop immediately
	netx := &netxlite.Netx{}
	ip, err := stunIPLookup(ctx, stunConfig{
		Endpoint: "stun.ekiga.net:3478",
		Logger:   log.Log,
		Resolver: netx.NewStdlibResolver(model.DiscardLogger),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the IP address we expected: %+v", ip)
	}
}

func TestSTUNIPLookupDialFailure(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	ip, err := stunIPLookup(ctx, stunConfig{
		Dialer: &mocks.Dialer{
			MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return nil, expected
			},
		},
		Endpoint: "stun.ekiga.net:3478",
		Logger:   log.Log,
	})
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the IP address we expected: %+v", ip)
	}
}

type MockableSTUNClient struct {
	StartErr error
	Event    stun.Event
}

func (c MockableSTUNClient) Close() error {
	return nil
}

func (c MockableSTUNClient) Start(m *stun.Message, h stun.Handler) error {
	if c.StartErr != nil {
		return c.StartErr
	}
	go func() {
		<-time.After(100 * time.Millisecond)
		h(c.Event)
	}()
	return nil
}

func TestSTUNIPLookupStartReturnsError(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	ip, err := stunIPLookup(ctx, stunConfig{
		Dialer: &mocks.Dialer{
			MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				conn := &mocks.Conn{}
				return conn, nil
			},
		},
		Endpoint: "stun.ekiga.net:3478",
		Logger:   log.Log,
		NewClient: func(conn net.Conn) (stunClient, error) {
			return MockableSTUNClient{StartErr: expected}, nil
		},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the IP address we expected: %+v", ip)
	}
}

func TestSTUNIPLookupStunEventContainsError(t *testing.T) {
	expected := errors.New("mocked error")
	ctx := context.Background()
	ip, err := stunIPLookup(ctx, stunConfig{
		Dialer: &mocks.Dialer{
			MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				conn := &mocks.Conn{}
				return conn, nil
			},
		},
		Endpoint: "stun.ekiga.net:3478",
		Logger:   log.Log,
		NewClient: func(conn net.Conn) (stunClient, error) {
			return MockableSTUNClient{Event: stun.Event{
				Error: expected,
			}}, nil
		},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the IP address we expected: %+v", ip)
	}
}

func TestSTUNIPLookupCannotDecodeMessage(t *testing.T) {
	ctx := context.Background()
	ip, err := stunIPLookup(ctx, stunConfig{
		Dialer: &mocks.Dialer{
			MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				conn := &mocks.Conn{}
				return conn, nil
			},
		},
		Endpoint: "stun.ekiga.net:3478",
		Logger:   log.Log,
		NewClient: func(conn net.Conn) (stunClient, error) {
			return MockableSTUNClient{Event: stun.Event{
				Message: &stun.Message{},
			}}, nil
		},
	})
	if !errors.Is(err, stun.ErrAttributeNotFound) {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if ip != model.DefaultProbeIP {
		t.Fatalf("not the IP address we expected: %+v", ip)
	}
}

func TestIPLookupWorksUsingSTUNEkiga(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	netx := &netxlite.Netx{}
	ip, err := stunEkigaIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		model.HTTPHeaderUserAgent,
		netx.NewStdlibResolver(model.DiscardLogger),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}

func TestIPLookupWorksUsingSTUNGoogle(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}

	netx := &netxlite.Netx{}
	ip, err := stunGoogleIPLookup(
		context.Background(),
		http.DefaultClient,
		log.Log,
		model.HTTPHeaderUserAgent,
		netx.NewStdlibResolver(model.DiscardLogger),
	)
	if err != nil {
		t.Fatal(err)
	}
	if net.ParseIP(ip) == nil {
		t.Fatalf("not an IP address: '%s'", ip)
	}
}
