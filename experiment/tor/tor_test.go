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
	"github.com/ooni/probe-engine/internal/orchestra"
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

func TestUnitMeasurerMeasureNewOrchestraClientError(t *testing.T) {
	measurer := newMeasurer(Config{})
	expected := errors.New("mocked error")
	measurer.newOrchestraClient = func(ctx context.Context, sess *session.Session) (*orchestra.Client, error) {
		return nil, expected
	}
	err := measurer.measure(
		context.Background(),
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMeasurerMeasureFetchTorTargetsError(t *testing.T) {
	measurer := newMeasurer(Config{})
	expected := errors.New("mocked error")
	measurer.newOrchestraClient = func(ctx context.Context, sess *session.Session) (*orchestra.Client, error) {
		return new(orchestra.Client), nil
	}
	measurer.fetchTorTargets = func(ctx context.Context, clnt *orchestra.Client) (map[string]model.TorTarget, error) {
		return nil, expected
	}
	err := measurer.measure(
		context.Background(),
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMeasurerMeasureGood(t *testing.T) {
	measurer := newMeasurer(Config{})
	measurer.newOrchestraClient = func(ctx context.Context, sess *session.Session) (*orchestra.Client, error) {
		return new(orchestra.Client), nil
	}
	measurer.fetchTorTargets = func(ctx context.Context, clnt *orchestra.Client) (map[string]model.TorTarget, error) {
		return nil, nil
	}
	err := measurer.measure(
		context.Background(),
		&session.Session{
			Logger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationMeasurerMeasureGood(t *testing.T) {
	measurer := newMeasurer(Config{})
	sess := session.New(
		log.Log,
		"ooniprobe-engine",
		"0.1.0-dev",
		"../../testdata/",
		nil, nil,
		"../../testdata/",
		kvstore.NewMemoryKeyValueStore(),
	)
	err := measurer.measure(
		context.Background(),
		sess,
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
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
		Address:  "1.1.1.1:80",
		Protocol: "tcp",
	},
}

func TestUnitMeasurerMeasureTargetsNoInput(t *testing.T) {
	var measurement model.Measurement
	measurer := new(measurer)
	measurer.measureTargets(
		context.Background(),
		&session.Session{
			Logger: log.Log,
		},
		&measurement,
		handler.NewPrinterCallbacks(log.Log),
		nil,
	)
	if len(measurement.TestKeys.(*TestKeys).Targets) != 0 {
		t.Fatal("expected no measurements here")
	}
}

func TestUnitMeasurerMeasureTargetsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we don't actually do anything
	var measurement model.Measurement
	measurer := new(measurer)
	measurer.measureTargets(
		ctx,
		&session.Session{
			Logger: log.Log,
		},
		&measurement,
		handler.NewPrinterCallbacks(log.Log),
		map[string]model.TorTarget{
			"xx": staticTestingTargets[0],
		},
	)
	targets := measurement.TestKeys.(*TestKeys).Targets
	if len(targets) != 1 {
		t.Fatal("expected single measurements here")
	}
	if _, found := targets["xx"]; !found {
		t.Fatal("the target we expected is missing")
	}
	tgt := targets["xx"]
	if !strings.HasSuffix(*tgt.Failure, "operation was canceled") {
		t.Fatal("not the error we expected")
	}
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
		context.Background(), keytarget{
			key:    "xx", // using an super simple key; should work anyway
			target: staticTestingTargets[0],
		},
		len(staticTestingTargets),
	)
	if len(rc.targetresults) != 1 {
		t.Fatal("wrong number of entries")
	}
	// Implementation note: here we won't bother with checking that
	// oonidatamodel works correctly because we already test that.
	if rc.targetresults["xx"].Agent != "redirect" {
		t.Fatal("agent is invalid")
	}
	if rc.targetresults["xx"].Failure != nil {
		t.Fatal("failure is invalid")
	}
	if rc.targetresults["xx"].TargetAddress != staticTestingTargets[0].Address {
		t.Fatal("target address is invalid")
	}
	if rc.targetresults["xx"].TargetProtocol != staticTestingTargets[0].Protocol {
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
		context.Background(), keytarget{
			key:    "xx", // using an super simple key; should work anyway
			target: staticTestingTargets[0],
		},
		len(staticTestingTargets),
	)
	if len(rc.targetresults) != 1 {
		t.Fatal("wrong number of entries")
	}
	// Implementation note: here we won't bother with checking that
	// oonidatamodel works correctly because we already test that.
	if rc.targetresults["xx"].Agent != "redirect" {
		t.Fatal("agent is invalid")
	}
	if *rc.targetresults["xx"].Failure != "mocked error" {
		t.Fatal("failure is invalid")
	}
	if rc.targetresults["xx"].TargetAddress != staticTestingTargets[0].Address {
		t.Fatal("target address is invalid")
	}
	if rc.targetresults["xx"].TargetProtocol != staticTestingTargets[0].Protocol {
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
	if !strings.HasSuffix(err.Error(), "operation was canceled") &&
		!strings.HasSuffix(err.Error(), "context canceled") {
		t.Fatalf("not the error we expected: %+v", err)
	}
	if tk.Connects == nil {
		t.Fatalf("expected connects data here, found: %+v", tk.Connects)
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
