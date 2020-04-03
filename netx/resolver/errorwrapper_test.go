package resolver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/netx/resolver"
)

func TestUnitErrorWrapperSuccess(t *testing.T) {
	orig := []string{"8.8.8.8"}
	r := resolver.ErrorWrapper{
		Resolver: resolver.NewMockableResolverWithResult(orig),
	}
	addrs, err := r.LookupHost(context.Background(), "dns.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != len(orig) || addrs[0] != orig[0] {
		t.Fatal("not the result we expected")
	}
}

func TestUnitErrorWrapperFailure(t *testing.T) {
	r := resolver.ErrorWrapper{
		Resolver: resolver.NewMockableResolverThatFails(),
	}
	ctx := context.Background()
	ctx = dialid.WithDialID(ctx)
	ctx = transactionid.WithTransactionID(ctx)
	addrs, err := r.LookupHost(ctx, "dns.google.com")
	if addrs != nil {
		t.Fatal("expected nil addr here")
	}
	var errWrapper *modelx.ErrWrapper
	if !errors.As(err, &errWrapper) {
		t.Fatal("cannot properly cast the returned error")
	}
	if errWrapper.Failure != modelx.FailureDNSNXDOMAINError {
		t.Fatal("unexpected failure")
	}
	if errWrapper.ConnID != 0 {
		t.Fatal("unexpected ConnID")
	}
	if errWrapper.DialID == 0 {
		t.Fatal("unexpected DialID")
	}
	if errWrapper.TransactionID == 0 {
		t.Fatal("unexpected TransactionID")
	}
	if errWrapper.Operation != "resolve" {
		t.Fatal("unexpected Operation")
	}
}
