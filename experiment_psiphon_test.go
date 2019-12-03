// +build !nopsiphon

package engine

import "testing"

func TestRunPsiphon(t *testing.T) {
	sess := newSessionForTesting(t)
	builder, err := sess.NewExperimentBuilder("psiphon")
	if err != nil {
		t.Fatal(err)
	}
	runexperimentflow(t, builder.Build())
}
