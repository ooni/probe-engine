// Package model2 contains stuff that cannot go into model because of
// circular dependencies. We will eventually fix this import loop such
// that there's no need to split the model package.
package model2

import (
	"context"

	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

// ExperimentMeasurer is the interface that allows to run a
// measurement for a specific experiment.
//
// This code cannot be in model because session depens on
// model, and we depend on session. We'll fix that.
type ExperimentMeasurer interface {
	// ExperimentName returns the experiment name.
	ExperimentName() string

	// ExperimentVersion returns the experiment version.
	ExperimentVersion() string

	// Run runs the experiment with the specified context, session,
	// measurement, and experiment calbacks.
	Run(
		ctx context.Context, sess *session.Session,
		measurement *model.Measurement, callbacks handler.Callbacks,
	) error
}
