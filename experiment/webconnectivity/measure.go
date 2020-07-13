package webconnectivity

import (
	"context"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
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

// DNSExperimentFailure returns the tk.DNSExperimentFailure value
// depending on the value of tk.Failure and tk.FailedOperation.
func DNSExperimentFailure(tk *TestKeys) (out *string) {
	if tk.Failure != nil && tk.FailedOperation != nil {
		switch *tk.FailedOperation {
		case modelx.ResolveOperation:
			out = tk.Failure
		}
	}
	return
}

// HTTPExperimentFailure return the tk.HTTPExperimentFailure value
// depending on the value of tk.Failure and tk.FailedOperation.
func HTTPExperimentFailure(tk *TestKeys) (out *string) {
	if tk.Failure != nil && tk.FailedOperation != nil {
		switch *tk.FailedOperation {
		case modelx.ConnectOperation, modelx.TLSHandshakeOperation,
			modelx.HTTPRoundTripOperation:
			out = tk.Failure
		}
	}
	return
}
