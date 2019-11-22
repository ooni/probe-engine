package bouncer_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/bouncer"
)

func makeClient() *bouncer.Client {
	return &bouncer.Client{
		BaseURL:    "https://ps-test.ooni.io/",
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}
}

func TestGetCollectors(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	collectors, err := makeClient().GetCollectors(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("%+v", collectors)
}

func TestGetTestHelpers(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	testhelpers, err := makeClient().GetTestHelpers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("%+v", testhelpers)
}
