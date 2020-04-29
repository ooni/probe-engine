package httptransport_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
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

func TestUnitSaverBodySuccess(t *testing.T) {
	saver := new(trace.Saver)
	txp := httptransport.SaverBodyHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Func: func(req *http.Request) (*http.Response, error) {
				data, err := ioutil.ReadAll(req.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != "deadbeef" {
					t.Fatal("invalid data")
				}
				return &http.Response{
					StatusCode: 501,
					Body:       ioutil.NopCloser(strings.NewReader("abad1dea")),
				}, nil
			},
		},
		SnapshotSize: 4,
		Saver:        saver,
	}
	body := strings.NewReader("deadbeef")
	req, err := http.NewRequest("POST", "http://x.org/y", body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 501 {
		t.Fatal("unexpected status code")
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "abad1dea" {
		t.Fatal("unexpected body")
	}
	ev := saver.Read()
	if len(ev) != 2 {
		t.Fatal("unexpected number of events")
	}
	if string(ev[0].Data) != "dead" {
		t.Fatal("invalid Data")
	}
	if ev[0].DataIsTruncated != true {
		t.Fatal("invalid DataIsTruncated")
	}
	if ev[0].Name != "http_request_body_snapshot" {
		t.Fatal("invalid Name")
	}
	if ev[0].Time.After(time.Now()) {
		t.Fatal("invalid Time")
	}
	if string(ev[1].Data) != "abad" {
		t.Fatal("invalid Data")
	}
	if ev[1].DataIsTruncated != true {
		t.Fatal("invalid DataIsTruncated")
	}
	if ev[1].Name != "http_response_body_snapshot" {
		t.Fatal("invalid Name")
	}
	if ev[1].Time.Before(ev[0].Time) {
		t.Fatal("invalid Time")
	}
}

func TestUnitSaverBodyRequestReadError(t *testing.T) {
	saver := new(trace.Saver)
	txp := httptransport.SaverBodyHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Func: func(req *http.Request) (*http.Response, error) {
				panic("should not be called")
			},
		},
		SnapshotSize: 4,
		Saver:        saver,
	}
	expected := errors.New("mocked error")
	body := httptransport.FakeBody{Err: expected}
	req, err := http.NewRequest("POST", "http://x.org/y", body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
	ev := saver.Read()
	if len(ev) != 0 {
		t.Fatal("unexpected number of events")
	}
}

func TestUnitSaverBodyRoundTripError(t *testing.T) {
	saver := new(trace.Saver)
	expected := errors.New("mocked error")
	txp := httptransport.SaverBodyHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Err: expected,
		},
		SnapshotSize: 4,
		Saver:        saver,
	}
	body := strings.NewReader("deadbeef")
	req, err := http.NewRequest("POST", "http://x.org/y", body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
	ev := saver.Read()
	if len(ev) != 1 {
		t.Fatal("unexpected number of events")
	}
	if string(ev[0].Data) != "dead" {
		t.Fatal("invalid Data")
	}
	if ev[0].DataIsTruncated != true {
		t.Fatal("invalid DataIsTruncated")
	}
	if ev[0].Name != "http_request_body_snapshot" {
		t.Fatal("invalid Name")
	}
	if ev[0].Time.After(time.Now()) {
		t.Fatal("invalid Time")
	}
}

func TestUnitSaverBodyResponseReadError(t *testing.T) {
	saver := new(trace.Saver)
	expected := errors.New("mocked error")
	txp := httptransport.SaverBodyHTTPTransport{
		RoundTripper: httptransport.FakeTransport{
			Func: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body: httptransport.FakeBody{
						Err: expected,
					},
				}, nil
			},
		},
		SnapshotSize: 4,
		Saver:        saver,
	}
	body := strings.NewReader("deadbeef")
	req, err := http.NewRequest("POST", "http://x.org/y", body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := txp.RoundTrip(req)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if resp != nil {
		t.Fatal("expected nil response")
	}
	ev := saver.Read()
	if len(ev) != 1 {
		t.Fatal("unexpected number of events")
	}
	if string(ev[0].Data) != "dead" {
		t.Fatal("invalid Data")
	}
	if ev[0].DataIsTruncated != true {
		t.Fatal("invalid DataIsTruncated")
	}
	if ev[0].Name != "http_request_body_snapshot" {
		t.Fatal("invalid Name")
	}
	if ev[0].Time.After(time.Now()) {
		t.Fatal("invalid Time")
	}
}
