package dialid

import (
	"context"
	"sync/atomic"
)

type contextkey struct{}

var id int64

// WithDialID returns a copy of ctx with DialID
func WithDialID(ctx context.Context) context.Context {
	return context.WithValue(
		ctx, contextkey{}, atomic.AddInt64(&id, 1),
	)
}

// ContextDialID returns the DialID of the context, or zero
func ContextDialID(ctx context.Context) int64 {
	id, _ := ctx.Value(contextkey{}).(int64)
	return id
}
