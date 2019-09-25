package session_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/session"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	sess := session.New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
	)

	sess.AvailableBouncers = append(sess.AvailableBouncers, model.Service{
		Address: "http://foobar.onion",
		Type:    "onion",
	})
	sess.AddAvailableHTTPSBouncer("https://bouncer.ooni.io")
	if len(sess.AvailableBouncers) != 2 {
		t.Fatal("unexpected size of available bouncers")
	}

	sess.AvailableCollectors = append(sess.AvailableCollectors, model.Service{
		Address: "http://foobar.onion",
		Type:    "onion",
	})
	sess.AddAvailableHTTPSCollector("https://b.collector.ooni.io")
	if len(sess.AvailableCollectors) != 2 {
		t.Fatal("unexpected size of available collectors")
	}

	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	readfile := func(path string) (err error) {
		_, err = ioutil.ReadFile(path)
		return
	}
	if err := readfile(sess.ASNDatabasePath()); err != nil {
		t.Fatal(err)
	}
	if err := readfile(sess.CABundlePath()); err != nil {
		t.Fatal(err)
	}
	if err := readfile(sess.CountryDatabasePath()); err != nil {
		t.Fatal(err)
	}

	// Repeating because the second lookup should be idempotent
	if err := sess.MaybeLookupTestHelpers(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestBouncerError(t *testing.T) {
	// Combine proxy testing with a broken proxy with errors
	// in reaching out to the bouncer.

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		},
	))
	defer server.Close()
	URL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	sess := session.New(
		log.Log, softwareName, softwareVersion, "../testdata", URL, nil,
	)

	if err := sess.MaybeLookupBackends(ctx); err == nil {
		t.Fatal("expected an error here")
	}
	if err := sess.MaybeLookupTestHelpers(ctx); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestLookupLocationError(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cause operations to fail
	sess := session.New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
	)
	if err := sess.MaybeLookupLocation(ctx); err == nil {
		t.Fatal("expected an error here")
	}
}
