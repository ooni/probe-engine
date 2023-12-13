package dslvm

import "github.com/ooni/probe-engine/pkg/model"

// Closer is something that [Drop] should explicitly close.
type Closer interface {
	Close(logger model.Logger) error
}
