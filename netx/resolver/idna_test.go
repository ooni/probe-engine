package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/netx/resolver"
)

var ErrUnexpectedPunycode = errors.New("unexpected punycode value")

type CheckIDNAResolver struct {
	Addresses []string
	Error     error
	Expect    string
}

// LookupHost implements Resolver.LookupHost
func (resolv CheckIDNAResolver) LookupHost(ctx context.Context,
	hostname string) ([]string, error) {
	if hostname != resolv.Expect {
		return nil, ErrUnexpectedPunycode
	}
	if resolv.Error != nil {
		return nil, resolv.Error
	}
	return resolv.Addresses, nil
}

// Network returns the transport network (e.g., doh, dot)
func (r CheckIDNAResolver) Network() string {
	return "checkidna"
}

// Address returns the upstream server address.
func (r CheckIDNAResolver) Address() string {
	return ""
}

func TestIDNAResolverSuccess(t *testing.T) {
	expectedAddr := "77.88.55.66"
	resolv := resolver.IDNAResolver{
		Resolver: CheckIDNAResolver{
			Addresses: []string{"5.255.255.55", "77.88.55.55",
				"77.88.55.66", "5.255.255.5", "2a02:6b8:a::a",
			},
			Expect: "xn--d1acpjx3f.xn--p1ai",
		},
	}
	addrs, err := resolv.LookupHost(context.Background(), "яндекс.рф")
	if err != nil {
		t.Fatal(err)
	}
	addrsMap := make(map[string]bool)
	for _, val := range addrs {
		addrsMap[val] = true
	}
	if _, ok := addrsMap[expectedAddr]; !ok {
		t.Fatal("IDNAResolver: unexpected differences")
	}
}

func TestIDNAResolverFailure(t *testing.T) {
	resolv := resolver.IDNAResolver{
		Resolver: CheckIDNAResolver{
			Addresses: []string{"1.2.3.4"},
			Expect:    "nothing.invalid",
		},
	}
	// see https://git.io/JU0be
	addrs, err := resolv.LookupHost(context.Background(), "東京\uFF0Ejp")
	if err == nil {
		t.Fatal("IDNAResolver: expected an error here")
	}
	if err != ErrUnexpectedPunycode {
		t.Fatal("not the error we expected")
	}
	if addrs != nil {
		t.Fatal("expected no response here")
	}
}

func TestUnitIDNAResolverTransportOK(t *testing.T) {
	resolv := resolver.IDNAResolver{
		Resolver: CheckIDNAResolver{
			Addresses: []string{"1.2.3.4"},
			Expect:    "nothing.invalid",
		},
	}
	if resolv.Network() != "idna" {
		t.Fatal("invalid network")
	}
	if resolv.Address() != "" {
		t.Fatal("invalid address")
	}
}
