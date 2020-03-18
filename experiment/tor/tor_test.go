package tor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/model"
)

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "tor" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.1.0" {
		t.Fatal("unexpected version")
	}
}

func TestUnitMeasurerMeasureNewOrchestraClientError(t *testing.T) {
	measurer := newMeasurer(Config{})
	expected := errors.New("mocked error")
	measurer.newOrchestraClient = func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error) {
		return nil, expected
	}
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
	measurer.newOrchestraClient = func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error) {
		return new(orchestra.Client), nil
	}
	measurer.fetchTorTargets = func(ctx context.Context, clnt model.ExperimentOrchestraClient) (map[string]model.TorTarget, error) {
		return nil, expected
	}
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
}

func TestUnitMeasurerMeasureFetchTorTargetsEmptyList(t *testing.T) {
	measurer := newMeasurer(Config{})
	measurer.newOrchestraClient = func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error) {
		return new(orchestra.Client), nil
	}
	measurer.fetchTorTargets = func(ctx context.Context, clnt model.ExperimentOrchestraClient) (map[string]model.TorTarget, error) {
		return nil, nil
	}
	measurement := new(model.Measurement)
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		measurement,
		handler.NewPrinterCallbacks(log.Log),
	)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*TestKeys)
	if len(tk.Targets) != 0 {
		t.Fatal("expected no targets here")
	}
}

func TestUnitMeasurerMeasureGood(t *testing.T) {
	measurer := newMeasurer(Config{})
	measurer.newOrchestraClient = func(ctx context.Context, sess model.ExperimentSession) (model.ExperimentOrchestraClient, error) {
		return new(orchestra.Client), nil
	}
	measurer.fetchTorTargets = func(ctx context.Context, clnt model.ExperimentOrchestraClient) (map[string]model.TorTarget, error) {
		return nil, nil
	}
	err := measurer.Run(
		context.Background(),
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
	sess := newsession()
	err := measurer.Run(
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
	model.TorTarget{
		Address: "192.95.36.142:443",
		Params: map[string][]string{
			"cert": []string{
				"qUVQ0srL1JI/vO6V6m/24anYXiJD3QP2HgzUKQtQ7GRqqUvs7P+tG43RtAqdhLOALP7DJQ",
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
	if rc.sentBytes.Load() != 10 {
		t.Fatal("sent bytes is invalid")
	}
	if rc.receivedBytes.Load() != 14 {
		t.Fatal("received bytes is invalid")
	}
}

func TestUnitResultsCollectorMeasureSingleTargetWithFailure(t *testing.T) {
	rc := newResultsCollector(
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
	if rc.sentBytes.Load() != 0 {
		t.Fatal("sent bytes is invalid")
	}
	if rc.receivedBytes.Load() != 0 {
		t.Fatal("received bytes is invalid")
	}
}

func TestUnitDefautFlexibleConnectDirPort(t *testing.T) {
	rc := newResultsCollector(
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
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

func TestUnitSummary(t *testing.T) {
	t.Run("without any piece of data", func(t *testing.T) {
		tr := new(TargetResults)
		tr.fillSummary()
		if len(tr.Summary) != 0 {
			t.Fatal("summary must be empty")
		}
	})

	t.Run("with a TCP connect and nothing else", func(t *testing.T) {
		tr := new(TargetResults)
		failure := "mocked_error"
		tr.TCPConnect = append(tr.TCPConnect, oonidatamodel.TCPConnectEntry{
			Status: oonidatamodel.TCPConnectStatus{
				Success: true,
				Failure: &failure,
			},
		})
		tr.fillSummary()
		if len(tr.Summary) != 1 {
			t.Fatal("cannot find expected entry")
		}
		if *tr.Summary["connect"].Failure != failure {
			t.Fatal("invalid failure")
		}
	})

	t.Run("for OBFS4", func(t *testing.T) {
		tr := new(TargetResults)
		tr.TCPConnect = append(tr.TCPConnect, oonidatamodel.TCPConnectEntry{
			Status: oonidatamodel.TCPConnectStatus{
				Success: true,
			},
		})
		failure := "mocked_error"
		tr.TargetProtocol = "obfs4"
		tr.Failure = &failure
		tr.fillSummary()
		if len(tr.Summary) != 2 {
			t.Fatal("cannot find expected entry")
		}
		if tr.Summary["connect"].Failure != nil {
			t.Fatal("invalid failure")
		}
		if *tr.Summary["handshake"].Failure != failure {
			t.Fatal("invalid failure")
		}
	})

	t.Run("for or_port/or_port_dirauth", func(t *testing.T) {
		doit := func(targetProtocol string, handshake *oonidatamodel.TLSHandshake) {
			tr := new(TargetResults)
			tr.TCPConnect = append(tr.TCPConnect, oonidatamodel.TCPConnectEntry{
				Status: oonidatamodel.TCPConnectStatus{
					Success: true,
				},
			})
			tr.TargetProtocol = targetProtocol
			if handshake != nil {
				tr.TLSHandshakes = append(tr.TLSHandshakes, *handshake)
			}
			tr.fillSummary()
			if len(tr.Summary) < 1 {
				t.Fatal("cannot find expected entry")
			}
			if tr.Summary["connect"].Failure != nil {
				t.Fatal("invalid failure")
			}
			if handshake == nil {
				if len(tr.Summary) != 1 {
					t.Fatal("unexpected summary length")
				}
				return
			}
			if len(tr.Summary) != 2 {
				t.Fatal("unexpected summary length")
			}
			if tr.Summary["handshake"].Failure != handshake.Failure {
				t.Fatal("the failure value is unexpected")
			}
		}
		doit("or_port_dirauth", nil)
		doit("or_port", nil)
	})
}

func TestUnitFillToplevelKeys(t *testing.T) {
	var tr TargetResults
	tr.TargetProtocol = "or_port"
	tk := new(TestKeys)
	tk.Targets = make(map[string]TargetResults)
	tk.Targets["xxx"] = tr
	tk.fillToplevelKeys()
	if tk.ORPortTotal != 1 {
		t.Fatal("unexpected ORPortTotal value")
	}
}

func newsession() model.ExperimentSession {
	return &mockable.ExperimentSession{MockableLogger: log.Log}
}
