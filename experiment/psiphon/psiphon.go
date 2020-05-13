// Package psiphon implements the psiphon network experiment. This
// implements, in particular, v0.2.0 of the spec.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-015-psiphon.md
package psiphon

import (
	"context"
	"sync"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
)

const (
	testName    = "psiphon"
	testVersion = "0.4.0"
)

// Config contains the experiment's configuration.
type Config struct {
	urlgetter.Config
}

// TestKeys contains the experiment's result.
type TestKeys struct {
	urlgetter.TestKeys
	MaxRuntime float64 `json:"max_runtime"`
}

// Measurer is the psiphon measurer.
type Measurer struct {
	BeforeGetHook func(g urlgetter.Getter)
	Config        Config
}

// ExperimentName returns the experiment name
func (m *Measurer) ExperimentName() string {
	return testName
}

// ExperimentVersion returns the experiment version
func (m *Measurer) ExperimentVersion() string {
	return testVersion
}

func (m *Measurer) printprogress(
	ctx context.Context, wg *sync.WaitGroup,
	maxruntime int, callbacks model.ExperimentCallbacks,
) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	step := 1 / float64(maxruntime)
	var progress float64
	defer callbacks.OnProgress(1.0, "psiphon experiment complete")
	defer wg.Done()
	for {
		select {
		case <-ticker.C:
			progress += step
			callbacks.OnProgress(progress, "psiphon experiment running")
		case <-ctx.Done():
			return
		}
	}
}

// Run runs the measurement
func (m *Measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	const maxruntime = 60
	ctx, cancel := context.WithTimeout(ctx, maxruntime*time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go m.printprogress(ctx, &wg, maxruntime, callbacks)
	m.Config.Tunnel = "psiphon" // force to use psiphon tunnel
	urlgetter.RegisterExtensions(measurement)
	target := "https://www.google.com/humans.txt"
	if measurement.Input != "" {
		target = string(measurement.Input)
	}
	g := urlgetter.Getter{
		Config:  m.Config.Config,
		Session: sess,
		Target:  target,
	}
	if m.BeforeGetHook != nil {
		m.BeforeGetHook(g)
	}
	tk, err := g.Get(ctx)
	cancel()
	wg.Wait()
	measurement.TestKeys = TestKeys{
		TestKeys:   tk,
		MaxRuntime: maxruntime,
	}
	return err
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return &Measurer{Config: config}
}
