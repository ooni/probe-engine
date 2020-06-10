package resolver

import (
	"context"

	"github.com/ooni/probe-engine/netx/errorx"
	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
	"github.com/ooni/probe-engine/netx/modelx"
)

// ErrorWrapperResolver is a Resolver that knows about wrapping errors.
type ErrorWrapperResolver struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost
func (r ErrorWrapperResolver) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	dialID := dialid.ContextDialID(ctx)
	txID := transactionid.ContextTransactionID(ctx)
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	err = errorx.SafeErrWrapperBuilder{
		DialID:        dialID,
		Error:         err,
		Operation:     modelx.ResolveOperation,
		TransactionID: txID,
	}.MaybeBuild()
	return addrs, err
}

var _ Resolver = ErrorWrapperResolver{}
