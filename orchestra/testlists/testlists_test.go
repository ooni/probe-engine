package testlists_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/orchestra/testlists"
)

func makeClient() *testlists.Client {
	return &testlists.Client{
		BaseURL:    testlists.DefaultBaseURL,
		HTTPClient: http.DefaultClient,
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
	}
}

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	urls, err := makeClient().Do(context.Background(), "IT")
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("%+v", urls)
}
