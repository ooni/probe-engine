// Package transactionid contains code to share the transactionID
package transactionid

import (
	"context"
	"sync/atomic"
)

type contextkey struct{}

var id int64

// WithTransactionID returns a copy of ctx with TransactionID
func WithTransactionID(ctx context.Context) context.Context {
	return context.WithValue(
		ctx, contextkey{}, atomic.AddInt64(&id, 1),
	)
}

// ContextTransactionID returns the TransactionID of the context, or zero
func ContextTransactionID(ctx context.Context) int64 {
	id, _ := ctx.Value(contextkey{}).(int64)
	return id
}
