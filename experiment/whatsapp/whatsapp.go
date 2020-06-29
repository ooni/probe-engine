// Package whatsapp contains the WhatsApp network experiment.
//
// See https://github.com/ooni/spec/blob/master/nettests/ts-018-whatsapp.md.
package whatsapp

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/model"
)

const (
	registrationServiceURL = "https://v.whatsapp.net/v2/register"
	testName               = "whatsapp"
	testVersion            = "0.7.0"
)

// Config contains the experiment config.
type Config struct {
	AllEndpoints bool `ooni:"Whether to test all WhatsApp endpoints"`
}

// TestKeys contains the experiment results
type TestKeys struct {
	urlgetter.TestKeys
}

// NewTestKeys returns a new instance of the test keys.
func NewTestKeys() *TestKeys {
	return new(TestKeys)
}

// Update updates the TestKeys using the given MultiOutput result.
func (tk *TestKeys) Update(v urlgetter.MultiOutput) {
	// update the easy to update entries first
	tk.NetworkEvents = append(tk.NetworkEvents, v.TestKeys.NetworkEvents...)
	tk.Queries = append(tk.Queries, v.TestKeys.Queries...)
	tk.Requests = append(tk.Requests, v.TestKeys.Requests...)
	tk.TCPConnect = append(tk.TCPConnect, v.TestKeys.TCPConnect...)
	tk.TLSHandshakes = append(tk.TLSHandshakes, v.TestKeys.TLSHandshakes...)
	// TODO(bassosimone): here we need to fill in all the fields that
	// are used by the WhatsApp expriment.
}

type measurer struct {
	config Config
}

func (m measurer) ExperimentName() string {
	return testName
}

func (m measurer) ExperimentVersion() string {
	return testVersion
}

func (m measurer) Run(
	ctx context.Context, sess model.ExperimentSession,
	measurement *model.Measurement, callbacks model.ExperimentCallbacks,
) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	urlgetter.RegisterExtensions(measurement)
	// generate all the inputs
	var inputs []urlgetter.MultiInput
	for idx := 1; idx <= 16; idx++ {
		for _, port := range []string{"443", "5222"} {
			inputs = append(inputs, urlgetter.MultiInput{
				Target: fmt.Sprintf("tcpconnect://e%d.whatsapp.net:%s", idx, port),
			})
		}
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(inputs), func(i, j int) {
		inputs[i], inputs[j] = inputs[j], inputs[i]
	})
	if m.config.AllEndpoints == false {
		inputs = inputs[0:1]
	}
	inputs = append(inputs, urlgetter.MultiInput{Target: registrationServiceURL})
	inputs = append(inputs, urlgetter.MultiInput{Target: "https://web.whatsapp.com"})
	inputs = append(inputs, urlgetter.MultiInput{Target: "http://web.whatsapp.com"})
	// measure in parallel
	multi := urlgetter.Multi{Begin: time.Now(), Session: sess}
	testkeys := NewTestKeys()
	testkeys.Agent = "redirect"
	measurement.TestKeys = testkeys
	for entry := range multi.Collect(ctx, inputs, "whatsapp", callbacks) {
		testkeys.Update(entry)
	}
	return nil
}

// NewExperimentMeasurer creates a new ExperimentMeasurer.
func NewExperimentMeasurer(config Config) model.ExperimentMeasurer {
	return measurer{config: config}
}
