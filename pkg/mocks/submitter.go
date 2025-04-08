package mocks

import (
	"context"

	"github.com/ooni/probe-engine/pkg/model"
)

// Submitter mocks model.Submitter.
type Submitter struct {
	MockSubmit func(ctx context.Context, m *model.Measurement) (string, error)
}

// Submit calls MockSubmit
func (s *Submitter) Submit(ctx context.Context, m *model.Measurement) (string, error) {
	return s.MockSubmit(ctx, m)
}
