package httptransport_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
)

func TestIntegrationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	counter := bytecounter.New()
	txp := httptransport.New(httptransport.Config{
		ByteCounter: counter,
		Logger:      log.Log,
	})
	client := &http.Client{Transport: txp}
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ioutil.ReadAll(resp.Body); err != nil {
		t.Fatal(err)
	}
	if err = resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log(counter.Sent.Load())
	t.Log(counter.Received.Load())
}
