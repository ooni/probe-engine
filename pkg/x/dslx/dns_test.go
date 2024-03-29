package dslx

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/mocks"
	"github.com/ooni/probe-engine/pkg/model"
)

/*
Test cases:
- New domain to resolve:
  - with empty domain
  - with options
*/
func TestNewDomainToResolve(t *testing.T) {
	t.Run("New domain to resolve", func(t *testing.T) {
		t.Run("with empty domain", func(t *testing.T) {
			domainToResolve := NewDomainToResolve(DomainName(""))
			if domainToResolve.Domain != "" {
				t.Fatalf("unexpected domain, want: %s, got: %s", "", domainToResolve.Domain)
			}
		})

		t.Run("with options", func(t *testing.T) {
			idGen := &atomic.Int64{}
			idGen.Add(42)
			domainToResolve := NewDomainToResolve(
				DomainName("www.example.com"),
				DNSLookupOptionTags("antani"),
			)
			if domainToResolve.Domain != "www.example.com" {
				t.Fatalf("unexpected domain")
			}
			if diff := cmp.Diff([]string{"antani"}, domainToResolve.Tags); diff != "" {
				t.Fatal(diff)
			}
		})
	})
}

/*
Test cases:
- Get dnsLookupGetaddrinfoFunc
- Apply dnsLookupGetaddrinfoFunc
  - with nil resolver
  - with lookup error
  - with success
*/
func TestGetaddrinfo(t *testing.T) {
	t.Run("Apply dnsLookupGetaddrinfoFunc", func(t *testing.T) {
		domain := &DomainToResolve{
			Domain: "example.com",
			Tags:   []string{"antani"},
		}

		t.Run("with nil resolver", func(t *testing.T) {
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now())
			f := DNSLookupGetaddrinfo(rt)
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // immediately cancel the lookup
			res := f.Apply(ctx, NewMaybeWithValue(domain))
			if obs := rt.Observations(); obs == nil || len(obs.Queries) <= 0 {
				t.Fatal("unexpected empty observations")
			}
			if res.Error == nil {
				t.Fatal("expected an error here")
			}
		})

		t.Run("with lookup error", func(t *testing.T) {
			mockedErr := errors.New("mocked")
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now(), RuntimeMeasurexLiteOptionMeasuringNetwork(&mocks.MeasuringNetwork{
				MockNewStdlibResolver: func(logger model.DebugLogger) model.Resolver {
					return &mocks.Resolver{
						MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
							return nil, mockedErr
						},
					}
				},
			}))
			f := DNSLookupGetaddrinfo(rt)
			res := f.Apply(context.Background(), NewMaybeWithValue(domain))
			if res.Error != mockedErr {
				t.Fatalf("unexpected error type: %s", res.Error)
			}
			if res.State != nil {
				t.Fatal("expected nil state")
			}
		})

		t.Run("with success", func(t *testing.T) {
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now(), RuntimeMeasurexLiteOptionMeasuringNetwork(&mocks.MeasuringNetwork{
				MockNewStdlibResolver: func(logger model.DebugLogger) model.Resolver {
					return &mocks.Resolver{
						MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
							return []string{"93.184.216.34"}, nil
						},
					}
				},
			}))
			f := DNSLookupGetaddrinfo(rt)
			res := f.Apply(context.Background(), NewMaybeWithValue(domain))
			if res.Error != nil {
				t.Fatalf("unexpected error: %s", res.Error)
			}
			if res.State == nil {
				t.Fatal("unexpected nil state")
			}
			if len(res.State.Addresses) != 1 || res.State.Addresses[0] != "93.184.216.34" {
				t.Fatal("unexpected addresses")
			}
		})
	})
}

/*
Test cases:
- Get dnsLookupUDPFunc
- Apply dnsLookupUDPFunc
  - with nil resolver
  - with lookup error
  - with success
*/
func TestLookupUDP(t *testing.T) {
	t.Run("Apply dnsLookupUDPFunc", func(t *testing.T) {
		domain := &DomainToResolve{
			Domain: "example.com",
			Tags:   []string{"antani"},
		}

		t.Run("with nil resolver", func(t *testing.T) {
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now())
			f := DNSLookupUDP(rt, "1.1.1.1:53")
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			res := f.Apply(ctx, NewMaybeWithValue(domain))
			if obs := rt.Observations(); obs == nil || len(obs.Queries) <= 0 {
				t.Fatal("unexpected empty observations")
			}
			if res.Error == nil {
				t.Fatalf("expected an error here")
			}
		})

		t.Run("with lookup error", func(t *testing.T) {
			mockedErr := errors.New("mocked")
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now(), RuntimeMeasurexLiteOptionMeasuringNetwork(&mocks.MeasuringNetwork{
				MockNewParallelUDPResolver: func(logger model.DebugLogger, dialer model.Dialer, endpoint string) model.Resolver {
					return &mocks.Resolver{
						MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
							return nil, mockedErr
						},
					}
				},
				MockNewDialerWithoutResolver: func(dl model.DebugLogger, w ...model.DialerWrapper) model.Dialer {
					return &mocks.Dialer{
						MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
							panic("should not be called")
						},
					}
				},
			}))
			f := DNSLookupUDP(rt, "1.1.1.1:53")
			res := f.Apply(context.Background(), NewMaybeWithValue(domain))
			if res.Error != mockedErr {
				t.Fatalf("unexpected error type: %s", res.Error)
			}
			if res.State != nil {
				t.Fatal("expected nil state")
			}
		})

		t.Run("with success", func(t *testing.T) {
			rt := NewRuntimeMeasurexLite(model.DiscardLogger, time.Now(), RuntimeMeasurexLiteOptionMeasuringNetwork(&mocks.MeasuringNetwork{
				MockNewParallelUDPResolver: func(logger model.DebugLogger, dialer model.Dialer, address string) model.Resolver {
					return &mocks.Resolver{
						MockLookupHost: func(ctx context.Context, domain string) ([]string, error) {
							return []string{"93.184.216.34"}, nil
						},
					}
				},
				MockNewDialerWithoutResolver: func(dl model.DebugLogger, w ...model.DialerWrapper) model.Dialer {
					return &mocks.Dialer{
						MockDialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
							panic("should not be called")
						},
					}
				},
			}))
			f := DNSLookupUDP(rt, "1.1.1.1:53")
			res := f.Apply(context.Background(), NewMaybeWithValue(domain))
			if res.Error != nil {
				t.Fatalf("unexpected error: %s", res.Error)
			}
			if res.State == nil {
				t.Fatal("unexpected nil state")
			}
			if len(res.State.Addresses) != 1 || res.State.Addresses[0] != "93.184.216.34" {
				t.Fatal("unexpected addresses")
			}
		})
	})
}
