package resolver

import (
	"context"

	"github.com/ooni/probe-engine/netx/internal/dialid"
	"github.com/ooni/probe-engine/netx/internal/errwrapper"
	"github.com/ooni/probe-engine/netx/internal/transactionid"
)

// ErrorWrapper is a Resolver that knows about wrapping errors.
type ErrorWrapper struct {
	Resolver
}

// LookupHost implements Resolver.LookupHost
func (r ErrorWrapper) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	dialID := dialid.ContextDialID(ctx)
	txID := transactionid.ContextTransactionID(ctx)
	addrs, err := r.Resolver.LookupHost(ctx, hostname)
	err = errwrapper.SafeErrWrapperBuilder{
		DialID:        dialID,
		Error:         err,
		Operation:     "resolve",
		TransactionID: txID,
	}.MaybeBuild()
	return addrs, err
}
