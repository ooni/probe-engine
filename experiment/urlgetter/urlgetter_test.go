package urlgetter_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/urlgetter"
	"github.com/ooni/probe-engine/internal/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
)

func TestMeasurer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := urlgetter.NewExperimentMeasurer(urlgetter.Config{})
	if m.ExperimentName() != "urlgetter" {
		t.Fatal("invalid experiment name")
	}
	if m.ExperimentVersion() != "0.0.3" {
		t.Fatal("invalid experiment version")
	}
	measurement := new(model.Measurement)
	measurement.Input = "https://www.google.com"
	err := m.Run(
		ctx, &mockable.ExperimentSession{},
		measurement, handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if len(measurement.Extensions) != 5 {
		t.Fatal("not the expected number of extensions")
	}
	tk := measurement.TestKeys.(urlgetter.TestKeys)
	if len(tk.DNSCache) != 0 {
		t.Fatal("not the DNSCache value we expected")
	}
}

func TestMeasurerDNSCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := urlgetter.NewExperimentMeasurer(urlgetter.Config{
		DNSCache: "dns.google 8.8.8.8 8.8.4.4",
	})
	if m.ExperimentName() != "urlgetter" {
		t.Fatal("invalid experiment name")
	}
	if m.ExperimentVersion() != "0.0.3" {
		t.Fatal("invalid experiment version")
	}
	measurement := new(model.Measurement)
	measurement.Input = "https://www.google.com"
	err := m.Run(
		ctx, &mockable.ExperimentSession{},
		measurement, handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if len(measurement.Extensions) != 5 {
		t.Fatal("not the expected number of extensions")
	}
	tk := measurement.TestKeys.(urlgetter.TestKeys)
	if len(tk.DNSCache) != 1 || tk.DNSCache[0] != "dns.google 8.8.8.8 8.8.4.4" {
		t.Fatal("invalid tk.DNSCache")
	}
}
