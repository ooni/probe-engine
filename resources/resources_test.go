package resources_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/httpx/httpx"
	"github.com/ooni/probe-engine/resources"
)

func TestEnsure(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	tempdir, err := ioutil.TempDir("", "ooniprobe-engine-resources-test")
	if err != nil {
		t.Fatal(err)
	}
	client := resources.Client{
		HTTPClient: httpx.NewTracingProxyingClient(log.Log, nil, nil),
		Logger:     log.Log,
		UserAgent:  "ooniprobe-engine/0.1.0",
		WorkDir:    tempdir,
	}
	err = client.Ensure(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// the second round should be idempotent
	err = client.Ensure(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
