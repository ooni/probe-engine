package webconnectivity

import (
	"context"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
)

// Measure performs the measurement given the context, the current
// measurement session, and the measurement target.
func Measure(
	ctx context.Context, sess model.ExperimentSession,
	target model.MeasurementTarget) (result urlgetter.TestKeys) {
	g := urlgetter.Getter{Session: sess, Target: string(target)}
	// Ignoring the error because g.Get() sets the tk.Failure field
	// to be the OONI equivalent of the error that occurred.
	result, _ = g.Get(ctx)
	return
}
