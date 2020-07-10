package hirl_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/apex/log"
	engine "github.com/ooni/probe-engine"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/experiment/hirl"
	"github.com/ooni/probe-engine/internal/mockable"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/modelx"
)

func TestNewExperimentMeasurer(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	if measurer.ExperimentName() != "http_invalid_request_line" {
		t.Fatal("unexpected name")
	}
	if measurer.ExperimentVersion() != "0.1.0" {
		t.Fatal("unexpected version")
	}
}

func TestIntegrationSuccess(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	ctx := context.Background()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t)
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*hirl.TestKeys)
	if len(tk.FailureList) != len(tk.Received) {
		t.Fatal("FailureList and Received have different lengths")
	}
	if len(tk.Received) != len(tk.Sent) {
		t.Fatal("Received and Sent have different lengths")
	}
	if len(tk.Sent) != len(tk.TamperingList) {
		t.Fatal("Sent and TamperingList have different lengths")
	}
	for _, failure := range tk.FailureList {
		if failure != nil {
			t.Fatal(*failure)
		}
	}
	for idx, received := range tk.Received {
		if received.Value != tk.Sent[idx] {
			t.Fatal("mismatch between received and sent")
		}
	}
	for _, entry := range tk.TamperingList {
		if entry != false {
			t.Fatal("found entry with tampering")
		}
	}
	if tk.Tampering != false {
		t.Fatal("overall there is tampering?!")
	}
}

func TestIntegrationCancelledContext(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// we need a real session because we need the tcp-echo helper
	sess := newsession(t)
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if err != nil {
		t.Fatal(err)
	}
	tk := measurement.TestKeys.(*hirl.TestKeys)
	if len(tk.FailureList) != 5 {
		t.Fatal("unexpected FailureList length")
	}
	for _, failure := range tk.FailureList {
		if *failure != modelx.FailureInterrupted {
			t.Fatal("unexpected failure")
		}
	}
	if len(tk.Received) != 5 {
		t.Fatal("unexpected Received length")
	}
	for _, entry := range tk.Received {
		if entry.Value != "" {
			t.Fatal("unexpected received entry")
		}
	}
	if len(tk.Sent) != 5 {
		t.Fatal("unexpected Sent length")
	}
	for _, entry := range tk.Sent {
		if entry != "" {
			t.Fatal("unexpected sent entry")
		}
	}
	if len(tk.TamperingList) != 5 {
		t.Fatal("unexpected TamperingList length")
	}
	for _, entry := range tk.TamperingList {
		if entry != false {
			t.Fatal("unexpected tampering entry")
		}
	}
	if tk.Tampering != false {
		t.Fatal("overall there is tampering?!")
	}
}

func TestNoHelpers(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	ctx := context.Background()
	sess := &mockable.ExperimentSession{}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, hirl.ErrNoAvailableTestHelpers) {
		t.Fatal("not the error we expected")
	}
	tk := measurement.TestKeys.(*hirl.TestKeys)
	if len(tk.FailureList) != 0 {
		t.Fatal("expected an empty FailureList")
	}
	if len(tk.FailureList) != len(tk.Received) {
		t.Fatal("FailureList and Received have different lengths")
	}
	if len(tk.Received) != len(tk.Sent) {
		t.Fatal("Received and Sent have different lengths")
	}
	if len(tk.Sent) != len(tk.TamperingList) {
		t.Fatal("Sent and TamperingList have different lengths")
	}
	if tk.Tampering != false {
		t.Fatal("overall there is tampering?!")
	}
}

func TestNoActualHelperInList(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	ctx := context.Background()
	sess := &mockable.ExperimentSession{
		MockableTestHelpers: map[string][]model.Service{
			"tcp-echo": nil,
		},
	}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, hirl.ErrNoAvailableTestHelpers) {
		t.Fatal("not the error we expected")
	}
	tk := measurement.TestKeys.(*hirl.TestKeys)
	if len(tk.FailureList) != 0 {
		t.Fatal("expected an empty FailureList")
	}
	if len(tk.FailureList) != len(tk.Received) {
		t.Fatal("FailureList and Received have different lengths")
	}
	if len(tk.Received) != len(tk.Sent) {
		t.Fatal("Received and Sent have different lengths")
	}
	if len(tk.Sent) != len(tk.TamperingList) {
		t.Fatal("Sent and TamperingList have different lengths")
	}
	if tk.Tampering != false {
		t.Fatal("overall there is tampering?!")
	}
}

func TestWrongTestHelperType(t *testing.T) {
	measurer := hirl.NewExperimentMeasurer(hirl.Config{})
	ctx := context.Background()
	sess := &mockable.ExperimentSession{
		MockableTestHelpers: map[string][]model.Service{
			"tcp-echo": {{
				Address: "127.0.0.1",
				Type:    "antani",
			}},
		},
	}
	measurement := new(model.Measurement)
	callbacks := handler.NewPrinterCallbacks(log.Log)
	err := measurer.Run(ctx, sess, measurement, callbacks)
	if !errors.Is(err, hirl.ErrInvalidHelperType) {
		t.Fatal("not the error we expected")
	}
	tk := measurement.TestKeys.(*hirl.TestKeys)
	if len(tk.FailureList) != 0 {
		t.Fatal("expected an empty FailureList")
	}
	if len(tk.FailureList) != len(tk.Received) {
		t.Fatal("FailureList and Received have different lengths")
	}
	if len(tk.Received) != len(tk.Sent) {
		t.Fatal("Received and Sent have different lengths")
	}
	if len(tk.Sent) != len(tk.TamperingList) {
		t.Fatal("Sent and TamperingList have different lengths")
	}
	if tk.Tampering != false {
		t.Fatal("overall there is tampering?!")
	}
}

func TestRunMethodDialFailure(t *testing.T) {
	sess := newsession(t)
	helpers, ok := sess.GetTestHelpersByName("tcp-echo")
	if len(helpers) < 1 || !ok {
		t.Fatal("cannot get helper")
	}
	expected := errors.New("mocked error")
	out := make(chan hirl.MethodResult)
	config := hirl.RunMethodConfig{
		MethodConfig: hirl.MethodConfig{
			Address: helpers[0].Address,
			Logger:  log.Log,
			Out:     out,
		},
		Name: "random_invalid_version_number",
		NewDialer: func(config httptransport.Config) httptransport.Dialer {
			return FakeDialer{Err: expected}
		},
		RequestLine: "GET / HTTP/ABC",
	}
	go hirl.RunMethod(context.Background(), config)
	result := <-out
	if !errors.Is(result.Err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Name != "random_invalid_version_number" {
		t.Fatal("unexpected Name")
	}
	if result.Received.Value != "" {
		t.Fatal("unexpected Received.Value")
	}
	if result.Sent != "" {
		t.Fatal("unexpected Sent")
	}
	if result.Tampering != false {
		t.Fatal("unexpected Tampering")
	}
}

func TestRunMethodSetDeadlineFailure(t *testing.T) {
	sess := newsession(t)
	helpers, ok := sess.GetTestHelpersByName("tcp-echo")
	if len(helpers) < 1 || !ok {
		t.Fatal("cannot get helper")
	}
	expected := errors.New("mocked error")
	out := make(chan hirl.MethodResult)
	config := hirl.RunMethodConfig{
		MethodConfig: hirl.MethodConfig{
			Address: helpers[0].Address,
			Logger:  log.Log,
			Out:     out,
		},
		Name: "random_invalid_version_number",
		NewDialer: func(config httptransport.Config) httptransport.Dialer {
			return FakeDialer{Conn: &FakeConn{
				SetDeadlineError: expected,
			}}
		},
		RequestLine: "GET / HTTP/ABC",
	}
	go hirl.RunMethod(context.Background(), config)
	result := <-out
	if !errors.Is(result.Err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Name != "random_invalid_version_number" {
		t.Fatal("unexpected Name")
	}
	if result.Received.Value != "" {
		t.Fatal("unexpected Received.Value")
	}
	if result.Sent != "" {
		t.Fatal("unexpected Sent")
	}
	if result.Tampering != false {
		t.Fatal("unexpected Tampering")
	}
}

func TestRunMethodWriteFailure(t *testing.T) {
	sess := newsession(t)
	helpers, ok := sess.GetTestHelpersByName("tcp-echo")
	if len(helpers) < 1 || !ok {
		t.Fatal("cannot get helper")
	}
	expected := errors.New("mocked error")
	out := make(chan hirl.MethodResult)
	config := hirl.RunMethodConfig{
		MethodConfig: hirl.MethodConfig{
			Address: helpers[0].Address,
			Logger:  log.Log,
			Out:     out,
		},
		Name: "random_invalid_version_number",
		NewDialer: func(config httptransport.Config) httptransport.Dialer {
			return FakeDialer{Conn: &FakeConn{
				WriteError: expected,
			}}
		},
		RequestLine: "GET / HTTP/ABC",
	}
	go hirl.RunMethod(context.Background(), config)
	result := <-out
	if !errors.Is(result.Err, expected) {
		t.Fatal("not the error we expected")
	}
	if result.Name != "random_invalid_version_number" {
		t.Fatal("unexpected Name")
	}
	if result.Received.Value != "" {
		t.Fatal("unexpected Received.Value")
	}
	if result.Sent != "" {
		t.Fatal("unexpected Sent")
	}
	if result.Tampering != false {
		t.Fatal("unexpected Tampering")
	}
}

func TestRunMethodReadEOFWithWrongData(t *testing.T) {
	sess := newsession(t)
	helpers, ok := sess.GetTestHelpersByName("tcp-echo")
	if len(helpers) < 1 || !ok {
		t.Fatal("cannot get helper")
	}
	out := make(chan hirl.MethodResult)
	config := hirl.RunMethodConfig{
		MethodConfig: hirl.MethodConfig{
			Address: helpers[0].Address,
			Logger:  log.Log,
			Out:     out,
		},
		Name: "random_invalid_version_number",
		NewDialer: func(config httptransport.Config) httptransport.Dialer {
			return FakeDialer{Conn: &FakeConn{
				ReadData: []byte("0xdeadbeef"),
			}}
		},
		RequestLine: "GET / HTTP/ABC",
	}
	go hirl.RunMethod(context.Background(), config)
	result := <-out
	if !errors.Is(result.Err, io.EOF) {
		t.Fatal("not the error we expected")
	}
	if result.Name != "random_invalid_version_number" {
		t.Fatal("unexpected Name")
	}
	if result.Received.Value != "0xdeadbeef" {
		t.Fatal("unexpected Received.Value")
	}
	if result.Sent != "GET / HTTP/ABC" {
		t.Fatal("unexpected Sent")
	}
	if result.Tampering != true {
		t.Fatal("unexpected Tampering")
	}
}

func newsession(t *testing.T) model.ExperimentSession {
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
	if err := sess.MaybeLookupBackends(); err != nil {
		t.Fatal(err)
	}
	return sess
}
