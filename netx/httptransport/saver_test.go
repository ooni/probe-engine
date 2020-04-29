package httptransport_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/dialer"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestUnitSaverPerformanceHTTPTransportSuccessContext(t *testing.T) {
	saver := &trace.Saver{}
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
				StatusCode: 200,
			},
		},
		Saver: saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately fail
	resp, err := txp.RoundTrip(req.WithContext(ctx))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
	ev := saver.Read()
	if len(ev) != 0 {
		t.Fatal("expected zero events")
	}
}

func TestIntegrationSaverPerformanceHTTPTransportSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: http.DefaultTransport.(*http.Transport),
		Saver:        saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non nil response here")
	}
	ev := saver.Read()
	if len(ev) != 3 {
		t.Fatal("expected two events")
	}
	//
	if ev[0].Name != "http_wrote_headers" {
		t.Fatal("unexpected Name")
	}
	if !ev[0].Time.Before(time.Now()) {
		t.Fatal("unexpected Time")
	}
	//
	if ev[1].Name != "http_wrote_request" {
		t.Fatal("unexpected Name")
	}
	if !ev[1].Time.After(ev[0].Time) {
		t.Fatal("unexpected Time")
	}
	//
	if ev[2].Name != "http_first_response_byte" {
		t.Fatal("unexpected Name")
	}
	if !ev[2].Time.After(ev[1].Time) {
		t.Fatal("unexpected Time")
	}
}

func TestIntegrationSaverPerformanceHTTPTransportSuccessByteCounting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	expCounter := bytecounter.New()
	sessCounter := bytecounter.New()
	ctx := context.Background()
	ctx = dialer.WithExperimentByteCounter(ctx, expCounter)
	ctx = dialer.WithSessionByteCounter(ctx, sessCounter)
	saver := &trace.Saver{}
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: httptransport.New(httptransport.Config{
			ContextByteCounting: true,
		}),
		Saver: saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req.WithContext(ctx))
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non nil response here")
	}
	if expCounter.Received.Load() <= 0 || expCounter.Sent.Load() <= 0 {
		t.Fatal("broken experiment counter")
	}
	if sessCounter.Received.Load() <= 0 || sessCounter.Received.Load() <= 0 {
		t.Fatal("broken session counter")
	}
}

func TestUnitSaverPerformanceHTTPTransportFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Err: expected,
		},
		Saver: saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
	ev := saver.Read()
	if len(ev) != 0 {
		t.Fatal("expected zero events")
	}
}

func TestUnitSaverPerformanceHTTPTransportFailureContext(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Err: expected,
		},
		Saver: saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cause immediate failure
	resp, err := txp.RoundTrip(req.WithContext(ctx))
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
	ev := saver.Read()
	if len(ev) != 0 {
		t.Fatal("expected zero events")
	}
}

func TestUnitSaverHTTPTransportFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	txp := httptransport.SaverHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Err: expected,
		},
		Saver: saver,
	}
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response here")
	}
	ev := saver.Read()
	if len(ev) != 2 {
		t.Fatal("expected two events")
	}
	if ev[0].HTTPRequest.Method != "GET" {
		t.Fatal("unexpected Method")
	}
	if ev[0].HTTPRequest.URL.String() != "http://www.google.com" {
		t.Fatal("unexpected URL")
	}
	if ev[0].Name != "http_round_trip_start" {
		t.Fatal("unexpected Name")
	}
	if !ev[0].Time.Before(time.Now()) {
		t.Fatal("unexpected Time")
	}
	if ev[1].Duration <= 0 {
		t.Fatal("unexpected Duration")
	}
	if !errors.Is(ev[1].Err, expected) {
		t.Fatal("unexpected Err")
	}
	if ev[1].HTTPRequest.Method != "GET" {
		t.Fatal("unexpected Method")
	}
	if ev[1].HTTPRequest.URL.String() != "http://www.google.com" {
		t.Fatal("unexpected URL")
	}
	if ev[1].HTTPResponse != nil {
		t.Fatal("unexpected HTTPResponse")
	}
	if ev[1].Name != "http_round_trip_done" {
		t.Fatal("unexpected Name")
	}
	if !ev[1].Time.After(ev[0].Time) {
		t.Fatal("unexpected Time")
	}
}

func TestIntegrationSaverHTTPTransportSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	txp := httptransport.SaverHTTPTransport{
		RoundTripper: http.DefaultTransport.(*http.Transport),
		Saver:        saver,
	}
	req, err := http.NewRequest("GET", "https://www.google.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if err != nil {
		t.Fatal("not the error we expected")
	}
	if resp == nil {
		t.Fatal("expected non nil response here")
	}
	ev := saver.Read()
	if len(ev) != 2 {
		t.Fatal("expected two events")
	}
	//
	if ev[0].HTTPRequest.Method != "GET" {
		t.Fatal("unexpected Method")
	}
	if ev[0].HTTPRequest.URL.String() != "https://www.google.com" {
		t.Fatal("unexpected URL")
	}
	if ev[0].Name != "http_round_trip_start" {
		t.Fatal("unexpected Name")
	}
	if !ev[0].Time.Before(time.Now()) {
		t.Fatal("unexpected Time")
	}
	//
	if ev[1].Duration <= 0 {
		t.Fatal("unexpected Duration")
	}
	if ev[1].Err != nil {
		t.Fatal("unexpected Err")
	}
	if ev[1].HTTPRequest.Method != "GET" {
		t.Fatal("unexpected Method")
	}
	if ev[1].HTTPRequest.URL.String() != "https://www.google.com" {
		t.Fatal("unexpected URL")
	}
	if ev[1].HTTPResponse.StatusCode != 200 {
		t.Fatal("unexpected StatusCode")
	}
	if ev[1].Name != "http_round_trip_done" {
		t.Fatal("unexpected Name")
	}
	if !ev[1].Time.After(ev[0].Time) {
		t.Fatal("unexpected Time")
	}
}
