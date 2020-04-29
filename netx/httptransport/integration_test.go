package httptransport_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/trace"
)

func TestIntegrationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	log.SetLevel(log.DebugLevel)
	counter := bytecounter.New()
	saver := new(trace.Saver)
	txp := httptransport.New(httptransport.Config{
		BogonIsError:        true,
		ByteCounter:         counter,
		CacheResolutions:    true,
		ContextByteCounting: true,
		Logger:              log.Log,
		SaveReadWrite:       true,
		Saver:               saver,
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
	if counter.Sent.Load() <= 0 {
		t.Fatal("no bytes sent?!")
	}
	if counter.Received.Load() <= 0 {
		t.Fatal("no bytes received?!")
	}
	if ev := saver.Read(); len(ev) <= 0 {
		t.Fatal("no low level events?!")
	}
}
