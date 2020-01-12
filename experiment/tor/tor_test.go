package tor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestUnitNewExperiment(t *testing.T) {
	sess := session.New(
		log.Log, softwareName, softwareVersion,
		"../../testdata", nil, nil, "../../testdata",
		kvstore.NewMemoryKeyValueStore(),
	)
	experiment := NewExperiment(sess, Config{})
	if experiment == nil {
		t.Fatal("nil experiment returned")
	}
}

func TestUnitMeasureWithCancelledContext(t *testing.T) {
	measurement := new(model.Measurement)
	m := newMeasurer(Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.measure(
		ctx,
		&session.Session{
			Logger: log.Log,
		},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range measurement.TestKeys.(*TestKeys).Targets {
		if entry.Failure == nil {
			t.Fatal("expected an error here")
		}
		good := (strings.HasSuffix(*entry.Failure, "operation was canceled") ||
			strings.HasSuffix(*entry.Failure, "context canceled"))
		if !good {
			t.Fatal("not the error we expected")
		}
	}
}

var staticTestingTargets = []model.TorTarget{
	// TODO(bassosimone): this is a public working bridge we have found
	// with @hellais. We should ask @phw whether there is some obfs4 bridge
	// dedicated to integration testing that we should use instead.
	model.TorTarget{
		Address: "109.105.109.165:10527",
		Params: map[string][]string{
			"cert": []string{
				"Bvg/itxeL4TWKLP6N1MaQzSOC6tcRIBv6q57DYAZc3b2AzuM+/TfB7mqTFEfXILCjEwzVA",
			},
			"iat-mode": []string{"1"},
		},
		Protocol: "obfs4",
	},
	model.TorTarget{
		Address:  "66.111.2.131:9030",
		Protocol: "dir_port",
	},
	model.TorTarget{
		Address:  "66.111.2.131:9001",
		Protocol: "or_port",
	},
	model.TorTarget{
		Address:  "www.google.com:80",
		Protocol: "tcp",
	},
}

func TestUnitResultsCollectorMeasureSingleTargetGood(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	rc.flexibleConnect = func(context.Context, model.TorTarget) (oonitemplates.Results, error) {
		return oonitemplates.Results{
			SentBytes:     10,
			ReceivedBytes: 14,
		}, nil
	}
	rc.measureSingleTarget(
		context.Background(), staticTestingTargets[0],
		len(staticTestingTargets),
	)
	if len(rc.targetresults) != 1 {
		t.Fatal("wrong number of entries")
	}
	// Implementation note: here we won't bother with checking that
	// oonidatamodel works correctly because we already test that.
	if rc.targetresults[0].Agent != "redirect" {
		t.Fatal("agent is invalid")
	}
	if rc.targetresults[0].Failure != nil {
		t.Fatal("failure is invalid")
	}
	if rc.targetresults[0].TargetAddress != staticTestingTargets[0].Address {
		t.Fatal("target address is invalid")
	}
	if rc.targetresults[0].TargetProtocol != staticTestingTargets[0].Protocol {
		t.Fatal("target protocol is invalid")
	}
	if rc.sentBytes != 10 {
		t.Fatal("sent bytes is invalid")
	}
	if rc.receivedBytes != 14 {
		t.Fatal("received bytes is invalid")
	}
}

func TestUnitResultsCollectorMeasureSingleTargetWithFailure(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	rc.flexibleConnect = func(context.Context, model.TorTarget) (oonitemplates.Results, error) {
		return oonitemplates.Results{}, errors.New("mocked error")
	}
	rc.measureSingleTarget(
		context.Background(), staticTestingTargets[0],
		len(staticTestingTargets),
	)
	if len(rc.targetresults) != 1 {
		t.Fatal("wrong number of entries")
	}
	// Implementation note: here we won't bother with checking that
	// oonidatamodel works correctly because we already test that.
	if rc.targetresults[0].Agent != "redirect" {
		t.Fatal("agent is invalid")
	}
	if *rc.targetresults[0].Failure != "mocked error" {
		t.Fatal("failure is invalid")
	}
	if rc.targetresults[0].TargetAddress != staticTestingTargets[0].Address {
		t.Fatal("target address is invalid")
	}
	if rc.targetresults[0].TargetProtocol != staticTestingTargets[0].Protocol {
		t.Fatal("target protocol is invalid")
	}
	if rc.sentBytes != 0 {
		t.Fatal("sent bytes is invalid")
	}
	if rc.receivedBytes != 0 {
		t.Fatal("received bytes is invalid")
	}
}

func TestUnitDefautFlexibleConnectDirPort(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tk, err := rc.defaultFlexibleConnect(ctx, staticTestingTargets[1])
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "context canceled") {
		t.Fatal("not the error we expected")
	}
	if tk.HTTPRequests == nil {
		t.Fatal("expected HTTP data here")
	}
}

func TestUnitDefautFlexibleConnectOrPort(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tk, err := rc.defaultFlexibleConnect(ctx, staticTestingTargets[2])
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
	if tk.Connects == nil {
		t.Fatal("expected connects data here")
	}
	if tk.NetworkEvents == nil {
		t.Fatal("expected network events data here")
	}
}

func TestUnitDefautFlexibleConnectOBFS4(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tk, err := rc.defaultFlexibleConnect(ctx, staticTestingTargets[0])
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
	if tk.Connects == nil {
		t.Fatal("expected connects data here")
	}
	if tk.NetworkEvents == nil {
		t.Fatal("expected network events data here")
	}
}

func TestUnitDefautFlexibleConnectDefault(t *testing.T) {
	rc := newResultsCollector(
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tk, err := rc.defaultFlexibleConnect(ctx, staticTestingTargets[3])
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "operation was canceled") {
		t.Fatal("not the error we expected")
	}
	if tk.Connects == nil {
		t.Fatal("expected connects data here")
	}
}

func TestUnitErrString(t *testing.T) {
	if errString(nil) != "success" {
		t.Fatal("not working with nil")
	}
	if errString(errors.New("antani")) != "antani" {
		t.Fatal("not working with error")
	}
}
