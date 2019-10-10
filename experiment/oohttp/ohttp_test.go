package oohttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestIntegration(t *testing.T) {
	mc := NewMeasuringClient(Config{})
	defer mc.Close()
	client := mc.HTTPClient()
	req, err := http.NewRequest("GET", "http://ooni.io", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	measurements := mc.PopMeasurementsByRoundTrip()
	if len(measurements) != 2 {
		t.Fatal("expected two round trips")
	}
}

func TestFailure(t *testing.T) {
	mc := NewMeasuringClient(Config{})
	defer mc.Close()
	mc.transport = &failingRoundTripper{}
	client := mc.HTTPClient()
	req, err := http.NewRequest("GET", "http://ooni.io", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if resp != nil {
		t.Fatal("expected a nil response here")
	}
}

type failingRoundTripper struct{}

func (rt *failingRoundTripper) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return nil, errors.New("mocked error")
}
