package tor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/internal/oonidatamodel"
	"github.com/ooni/probe-engine/internal/oonitemplates"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/modelx"
	"github.com/ooni/probe-engine/probeservices"
)

func TestUnitNewExperimentMeasurer(t *testing.T) {
	measurer := NewExperimentMeasurer(Config{})
	if measurer.ExperimentName() != "tor" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.2.0" {
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
		return new(probeservices.Client), nil
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
		return new(probeservices.Client), nil
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
		return new(probeservices.Client), nil
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
	{
		Address: "192.95.36.142:443",
		Params: map[string][]string{
			"cert": {
				"qUVQ0srL1JI/vO6V6m/24anYXiJD3QP2HgzUKQtQ7GRqqUvs7P+tG43RtAqdhLOALP7DJQ",
			},
			"iat-mode": {"1"},
		},
		Protocol: "obfs4",
	},
	{
		Address:  "66.111.2.131:9030",
		Protocol: "dir_port",
	},
	{
		Address:  "66.111.2.131:9001",
		Protocol: "or_port",
	},
	{
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

func wrapTestingTarget(tt model.TorTarget) keytarget {
	return keytarget{
		key:    "xx", // using an super simple key; should work anyway
		target: tt,
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
	rc.flexibleConnect = func(context.Context, keytarget) (oonitemplates.Results, error) {
		return oonitemplates.Results{}, nil
	}
	rc.measureSingleTarget(
		context.Background(), wrapTestingTarget(staticTestingTargets[0]),
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
}

func TestUnitResultsCollectorMeasureSingleTargetWithFailure(t *testing.T) {
	rc := newResultsCollector(
		&mockable.ExperimentSession{
			MockableLogger: log.Log,
		},
		new(model.Measurement),
		handler.NewPrinterCallbacks(log.Log),
	)
	rc.flexibleConnect = func(context.Context, keytarget) (oonitemplates.Results, error) {
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
	tk, err := rc.defaultFlexibleConnect(ctx, wrapTestingTarget(staticTestingTargets[1]))
	if err == nil {
		t.Fatal("expected an error here")
	}
	if !strings.HasSuffix(err.Error(), "interrupted") {
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
	tk, err := rc.defaultFlexibleConnect(ctx, wrapTestingTarget(staticTestingTargets[2]))
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
	tk, err := rc.defaultFlexibleConnect(ctx, wrapTestingTarget(staticTestingTargets[0]))
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
	tk, err := rc.defaultFlexibleConnect(ctx, wrapTestingTarget(staticTestingTargets[3]))
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
		if *tr.Summary[modelx.ConnectOperation].Failure != failure {
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
		if tr.Summary[modelx.ConnectOperation].Failure != nil {
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
			if tr.Summary[modelx.ConnectOperation].Failure != nil {
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
		doit("or_port", &oonidatamodel.TLSHandshake{
			Failure: (func() *string {
				s := io.EOF.Error()
				return &s
			})(),
		})
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
	return &mockable.ExperimentSession{
		MockableLogger:     log.Log,
		MockableHTTPClient: http.DefaultClient,
	}
}

var referenceTargetResult = []byte(`{
	"agent": "redirect",
	"failure": null,
	"network_events": [
	  {
		"address": "85.31.186.98:443",
		"conn_id": 19,
		"dial_id": 21,
		"failure": null,
		"operation": "connect",
		"proto": "tcp",
		"t": 8.639313
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1915,
		"operation": "write",
		"proto": "tcp",
		"t": 8.639686
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1440,
		"operation": "read",
		"proto": "tcp",
		"t": 8.691708
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1440,
		"operation": "read",
		"proto": "tcp",
		"t": 8.691912
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1383,
		"operation": "read",
		"proto": "tcp",
		"t": 8.69234
	  }
	],
	"queries": null,
	"requests": null,
	"summary": {
	  "connect": {
		"failure": null
	  }
	},
	"target_address": "85.31.186.98:443",
	"target_protocol": "obfs4",
	"tcp_connect": [
	  {
		"conn_id": 19,
		"dial_id": 21,
		"ip": "85.31.186.98",
		"port": 443,
		"status": {
		  "failure": null,
		  "success": true
		},
		"t": 8.639313
	  }
	],
	"tls_handshakes": null
  }`)

var scrubbedTargetResult = []byte(`{
	"agent": "redirect",
	"failure": null,
	"network_events": [
	  {
		"address": "[scrubbed]",
		"conn_id": 19,
		"dial_id": 21,
		"failure": null,
		"operation": "connect",
		"proto": "tcp",
		"t": 8.639313
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1915,
		"operation": "write",
		"proto": "tcp",
		"t": 8.639686
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1440,
		"operation": "read",
		"proto": "tcp",
		"t": 8.691708
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1440,
		"operation": "read",
		"proto": "tcp",
		"t": 8.691912
	  },
	  {
		"conn_id": 19,
		"failure": null,
		"num_bytes": 1383,
		"operation": "read",
		"proto": "tcp",
		"t": 8.69234
	  }
	],
	"queries": null,
	"requests": null,
	"summary": {
	  "connect": {
		"failure": null
	  }
	},
	"target_address": "[scrubbed]",
	"target_protocol": "obfs4",
	"tcp_connect": [
	  {
		"conn_id": 19,
		"dial_id": 21,
		"ip": "[scrubbed]",
		"port": 443,
		"status": {
		  "failure": null,
		  "success": true
		},
		"t": 8.639313
	  }
	],
	"tls_handshakes": null
  }`)

func TestMaybeSanitize(t *testing.T) {
	var input TargetResults
	if err := json.Unmarshal(referenceTargetResult, &input); err != nil {
		t.Fatal(err)
	}
	t.Run("nothing to do", func(t *testing.T) {
		out := maybeSanitize(input, keytarget{target: model.TorTarget{Source: ""}})
		diff := cmp.Diff(input, out)
		if diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("scrubbing to do", func(t *testing.T) {
		var expected TargetResults
		if err := json.Unmarshal(scrubbedTargetResult, &expected); err != nil {
			t.Fatal(err)
		}
		out := maybeSanitize(input, keytarget{target: model.TorTarget{
			Address: "85.31.186.98:443",
			Source:  "bridgedb",
		}})
		diff := cmp.Diff(expected, out)
		if diff != "" {
			t.Fatal(diff)
		}
	})
}

type savingLogger struct {
	debug []string
	info  []string
	warn  []string
}

func (sl *savingLogger) Debug(message string) {
	sl.debug = append(sl.debug, message)
}

func (sl *savingLogger) Debugf(format string, v ...interface{}) {
	sl.Debug(fmt.Sprintf(format, v...))
}

func (sl *savingLogger) Info(message string) {
	sl.info = append(sl.info, message)
}

func (sl *savingLogger) Infof(format string, v ...interface{}) {
	sl.Info(fmt.Sprintf(format, v...))
}

func (sl *savingLogger) Warn(message string) {
	sl.warn = append(sl.warn, message)
}

func (sl *savingLogger) Warnf(format string, v ...interface{}) {
	sl.Warn(fmt.Sprintf(format, v...))
}

func TestScrubLogger(t *testing.T) {
	input := "failure: 130.192.91.211:443: no route the host"
	expect := "failure: [scrubbed]: no route the host"

	t.Run("for debug", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Debug(input)
		if len(logger.debug) != 1 && len(logger.info) != 0 && len(logger.warn) != 0 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.debug[0] != expect {
			t.Fatal("unexpected output written")
		}
	})

	t.Run("for debugf", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Debugf("%s", input)
		if len(logger.debug) != 1 && len(logger.info) != 0 && len(logger.warn) != 0 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.debug[0] != expect {
			t.Fatal("unexpected output written")
		}
	})

	t.Run("for info", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Info(input)
		if len(logger.debug) != 0 && len(logger.info) != 1 && len(logger.warn) != 0 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.info[0] != expect {
			t.Fatal("unexpected output written")
		}
	})

	t.Run("for infof", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Infof("%s", input)
		if len(logger.debug) != 0 && len(logger.info) != 1 && len(logger.warn) != 0 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.info[0] != expect {
			t.Fatal("unexpected output written")
		}
	})

	t.Run("for warn", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Warn(input)
		if len(logger.debug) != 0 && len(logger.info) != 0 && len(logger.warn) != 1 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.warn[0] != expect {
			t.Fatal("unexpected output written")
		}
	})

	t.Run("for warnf", func(t *testing.T) {
		logger := new(savingLogger)
		scrubber := scrubbingLogger{Logger: logger}
		scrubber.Warnf("%s", input)
		if len(logger.debug) != 0 && len(logger.info) != 0 && len(logger.warn) != 1 {
			t.Fatal("unexpected number of log lines written")
		}
		if logger.warn[0] != expect {
			t.Fatal("unexpected output written")
		}
	})
}

func TestMaybeScrubbingLogger(t *testing.T) {
	var input model.Logger = new(savingLogger)

	t.Run("for when we don't need to save", func(t *testing.T) {
		kt := keytarget{target: model.TorTarget{
			Source: "",
		}}
		out := maybeScrubbingLogger(input, kt)
		if out != input {
			t.Fatal("not the output we expected")
		}
		if _, ok := out.(*savingLogger); !ok {
			t.Fatal("not the output type we expected")
		}
	})

	t.Run("for when we need to save", func(t *testing.T) {
		kt := keytarget{target: model.TorTarget{
			Source: "bridgedb",
		}}
		out := maybeScrubbingLogger(input, kt)
		if out == input {
			t.Fatal("not the output value we expected")
		}
		if _, ok := out.(scrubbingLogger); !ok {
			t.Fatal("not the output type we expected")
		}
	})
}
