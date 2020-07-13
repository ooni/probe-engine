package webconnectivity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/apex/log"
	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/webconnectivity"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	if measurer.ExperimentName() != "web_connectivity" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.1.0" {
		t.Fatal("unexpected version")
	}
}

func TestIntegrationSuccess(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	ctx := context.Background()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t, true)
	measurement := &model.Measurement{Input: "http://www.example.com"}
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*webconnectivity.TestKeys)
	if tk.Failure != nil {
		t.Fatal("unexpected failure")
	}
	if tk.ClientResolver == "" || tk.ClientResolver == model.DefaultResolverIP {
		t.Fatal("unexpected client_resolver")
	}
	if tk.ControlFailure != nil {
		t.Fatal("unexpected control_failure")
	}
	// TODO(bassosimone): write further checks here?
}

func TestMeasureWithCancelledContext(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t, true)
	measurement := &model.Measurement{Input: "http://www.example.com"}
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*webconnectivity.TestKeys)
	if *tk.Failure != modelx.FailureInterrupted {
		t.Fatal("unexpected failure")
	}
	if tk.ClientResolver == "" || tk.ClientResolver == model.DefaultResolverIP {
		t.Fatal("unexpected client_resolver")
	}
	if *tk.ControlFailure != modelx.FailureInterrupted {
		t.Fatal("unexpected control_failure")
	}
	// TODO(bassosimone): write further checks here?
}

func TestMeasureWithNoInput(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t, true)
	measurement := &model.Measurement{Input: ""}
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, webconnectivity.ErrNoInput) {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*webconnectivity.TestKeys)
	if tk.Failure != nil {
		t.Fatal("unexpected failure")
	}
	if tk.ClientResolver == "" || tk.ClientResolver == model.DefaultResolverIP {
		t.Fatal("unexpected client_resolver")
	}
	if tk.ControlFailure != nil {
		t.Fatal("unexpected control_failure")
	}
	// TODO(bassosimone): write further checks here?
}

func TestMeasureWithUnsupportedInput(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t, true)
	measurement := &model.Measurement{Input: "dnslookup://example.com"}
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, webconnectivity.ErrUnsupportedInput) {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*webconnectivity.TestKeys)
	if tk.Failure != nil {
		t.Fatal("unexpected failure")
	}
	if tk.ClientResolver == "" || tk.ClientResolver == model.DefaultResolverIP {
		t.Fatal("unexpected client_resolver")
	}
	if tk.ControlFailure != nil {
		t.Fatal("unexpected control_failure")
	}
	// TODO(bassosimone): write further checks here?
}

func TestMeasureWithNoAvailableTestHelpers(t *testing.T) {
	measurer := webconnectivity.NewExperimentMeasurer(webconnectivity.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t, false)
	measurement := &model.Measurement{Input: "https://www.example.com"}
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, webconnectivity.ErrNoAvailableTestHelpers) {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*webconnectivity.TestKeys)
	if tk.Failure != nil {
		t.Fatal("unexpected failure")
	}
	if tk.ClientResolver == "" || tk.ClientResolver == model.DefaultResolverIP {
		t.Fatal("unexpected client_resolver")
	}
	if tk.ControlFailure != nil {
		t.Fatal("unexpected control_failure")
	}
	// TODO(bassosimone): write further checks here?
}

func newsession(t *testing.T, lookupBackends bool) model.ExperimentSession {
	sess, err := engine.NewSession(engine.SessionConfig{
		AssetsDir: "../../testdata",
		AvailableProbeServices: []model.Service{{
			Address: "https://ps-test.ooni.io",
			Type:    "https",
		}},
		Logger: log.Log,
		PrivacySettings: model.PrivacySettings{
			IncludeASN:     true,
			IncludeCountry: true,
			IncludeIP:      false,
		},
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if lookupBackends {
		if err := sess.MaybeLookupBackends(); err != nil {
			t.Fatal(err)
		}
	}
	if err := sess.MaybeLookupLocation(); err != nil {
		t.Fatal(err)
	}
	return sess
}
