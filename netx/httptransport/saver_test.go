package httptransport_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestUnitSaverRoundTripFailure(t *testing.T) {
	expected := errors.New("mocked error")
	saver := &trace.Saver{}
	txp := httptransport.SaverRoundTripHTTPTransport{
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

func TestIntegrationSaverRoundTripSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	txp := httptransport.SaverRoundTripHTTPTransport{
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

func TestIntegrationSaverPerformanceNoMultipleEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := &trace.Saver{}
	// register twice - do we see events twice?
	txp := httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: http.DefaultTransport.(*http.Transport),
		Saver:        saver,
	}
	txp = httptransport.SaverPerformanceHTTPTransport{
		RoundTripper: txp,
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
	// we should specifically see the events not attached to any
	// context being submitted twice. This is fine because they are
	// explicit, while the context is implicit and hence leads to
	// more subtle bugs. For example, this happens when you measure
	// every event and combine HTTP with DoH.
	if len(ev) != 3 {
		t.Fatal("expected three events")
	}
	expected := []string{
		"http_wrote_headers",       // measured with context
		"http_wrote_request",       // measured with context
		"http_first_response_byte", // measured with context
	}
	for i := 0; i < len(expected); i++ {
		if ev[i].Name != expected[i] {
			t.Fatal("unexpected event name")
		}
	}
}
