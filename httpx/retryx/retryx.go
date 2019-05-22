// Package retryx helps to retry operations
package retryx

import (
	"context"
	"time"

	"github.com/avast/retry-go"
)

var (
	// Attempts is the number of attempts
	Attempts uint = 4

	// Delay is the base interval for exponential backoff
	Delay = 500 * time.Millisecond
)

// Do retries fn for Attempts times using exponentiat backoff with
// a base interval equal to Delay.
func Do(ctx context.Context, fn func() error) error {
	// Implementation note: not the optimal solution because the retry
	// package will sleep and we would like instead to always be interruptible
	// by the context. Yet, the maximum wait time with 4 attemtps and 500
	// millisecond is 1s, so it's not that bad. To be improved.
	return retry.Do(
		fn, retry.Attempts(Attempts), retry.Delay(Delay),
		retry.RetryIf(func(err error) bool {
			return ctx.Done() == nil
		}),
	)
}
