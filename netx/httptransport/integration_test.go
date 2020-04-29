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
	config := httptransport.Config{
		BogonIsError:        true,
		ByteCounter:         counter,
		CacheResolutions:    true,
		ContextByteCounting: true,
		DialSaver:           &trace.Saver{},
		HTTPSaver:           &trace.Saver{},
		Logger:              log.Log,
		ReadWriteSaver:      &trace.Saver{},
		ResolveSaver:        &trace.Saver{},
		TLSSaver:            &trace.Saver{},
	}
	txp := httptransport.New(config)
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
	if ev := config.DialSaver.Read(); len(ev) <= 0 {
		t.Fatal("no dial events?!")
	}
	if ev := config.HTTPSaver.Read(); len(ev) <= 0 {
		t.Fatal("no HTTP events?!")
	}
	if ev := config.ReadWriteSaver.Read(); len(ev) <= 0 {
		t.Fatal("no R/W events?!")
	}
	if ev := config.ResolveSaver.Read(); len(ev) <= 0 {
		t.Fatal("no resolver events?!")
	}
	if ev := config.TLSSaver.Read(); len(ev) <= 0 {
		t.Fatal("no TLS events?!")
	}
}
