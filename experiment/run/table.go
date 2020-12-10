package run

import (
	"context"

	"github.com/ooni/probe-engine/model"
)

type experimentMain func(ctx context.Context, input StructuredInput,
	sess model.ExperimentSession, measurement *model.Measurement,
	callbacks model.ExperimentCallbacks) error

var table = map[string]experimentMain{
	// TODO(bassosimone): before extending run to support more than
	// single experiment, we need to handle the case in which we are
	// including different experiments into the same report ID.
	// Probably, the right way to implement this functionality is to
	// use proveservices.Submitter to submit reports.
	"dnscheck": dodnscheck,
}
