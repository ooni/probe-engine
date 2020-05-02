package httptransport_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/netx/bytecounter"
	"github.com/ooni/probe-engine/netx/httptransport"
	"github.com/ooni/probe-engine/netx/modelx"
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

func TestIntegrationBogonResolutionNotBroken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	saver := new(trace.Saver)
	r := httptransport.NewResolver(httptransport.Config{
		BogonIsError: true,
		DNSCache: map[string][]string{
			"www.google.com": {"127.0.0.1"},
		},
		ResolveSaver: saver,
		Logger:       log.Log,
	})
	addrs, err := r.LookupHost(context.Background(), "www.google.com")
	if !errors.Is(err, modelx.ErrDNSBogon) {
		t.Fatal("not the error we expected")
	}
	if err.Error() != modelx.FailureDNSBogonError {
		t.Fatal("error not correctly wrapped")
	}
	if len(addrs) != 1 || addrs[0] != "127.0.0.1" {
		t.Fatal("address was not returned")
	}
}
