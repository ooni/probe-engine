package webconnectivitylte

//
// Summary
//

import "github.com/ooni/probe-engine/pkg/model"

// Summary contains the summary results.
//
// Note that this structure is part of the ABI contract with ooniprobe
// therefore we should be careful when changing it.
type SummaryKeys struct {
	// TODO: add here additional summary fields.
	isAnomaly bool
}

// GetSummaryKeys implements model.ExperimentMeasurer.GetSummaryKeys.
func (m *Measurer) GetSummaryKeys(measurement *model.Measurement) (any, error) {
	// TODO(bassosimone): fill all the SummaryKeys
	sk := SummaryKeys{isAnomaly: false}
	return sk, nil
}
