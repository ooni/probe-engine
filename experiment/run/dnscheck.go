package run

import (
	"context"

	"github.com/ooni/probe-engine/experiment/dnscheck"
	"github.com/ooni/probe-engine/model"
)

func dodnscheck(ctx context.Context, input StructuredInput,
	sess model.ExperimentSession, measurement *model.Measurement,
	callbacks model.ExperimentCallbacks) error {
	exp := dnscheck.Measurer{Config: input.DNSCheck}
	measurement.TestName = exp.ExperimentName()
	measurement.TestVersion = exp.ExperimentVersion()
	measurement.Input = model.MeasurementTarget(input.Input)
	return exp.Run(ctx, sess, measurement, callbacks)
}
