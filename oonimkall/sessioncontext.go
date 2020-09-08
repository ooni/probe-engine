package oonimkall

import (
	"context"
	"math"
	"time"
)

const maxTimeout = int64(time.Duration(math.MaxInt64) / time.Second)

func newContext(timeout int64) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		if timeout > maxTimeout {
			timeout = maxTimeout
		}
		return context.WithTimeout(
			context.Background(), time.Duration(timeout)*time.Second)
	}
	return context.WithCancel(context.Background())
}
