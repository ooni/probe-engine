package session

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/model"
)

const (
	softwareName    = "ooniprobe-example"
	softwareVersion = "0.0.1"
)

func TestIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)

	sess.AvailableBouncers = append(sess.AvailableBouncers, model.Service{
		Address: "http://foobar.onion",
		Type:    "onion",
	})
	sess.AddAvailableHTTPSBouncer("https://ps-test.ooni.io")
	if len(sess.AvailableBouncers) != 2 {
		t.Fatal("unexpected size of available bouncers")
	}

	sess.AvailableCollectors = append(sess.AvailableCollectors, model.Service{
		Address: "http://foobar.onion",
		Type:    "onion",
	})
	sess.AddAvailableHTTPSCollector("https://ps-test.ooni.io")
	if len(sess.AvailableCollectors) != 2 {
		t.Fatal("unexpected size of available collectors")
	}

	if err := sess.MaybeLookupBackends(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeLookupLocation(ctx); err != nil {
		t.Fatal(err)
	}

	if sess.ProbeASNString() == model.DefaultProbeASNString {
		t.Fatal("unexpected ProbeASNString")
	}
	if sess.ProbeASN() == model.DefaultProbeASN {
		t.Fatal("unexpected ProbeASN")
	}
	if sess.ProbeCC() == model.DefaultProbeCC {
		t.Fatal("unexpected ProbeCC")
	}
	if sess.ProbeIP() == model.DefaultProbeIP {
		t.Fatal("unexpected ProbeIP")
	}
	if sess.ProbeNetworkName() == model.DefaultProbeNetworkName {
		t.Fatal("unexpected ProbeNetworkName")
	}
	if sess.ResolverIP() == model.DefaultResolverIP {
		t.Fatal("unexpected ResolverIP")
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

func TestIntegrationNewOrchestraClient(t *testing.T) {
	ctx := context.Background()
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)
	clnt, err := sess.NewOrchestraClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if clnt == nil {
		t.Fatal("expected non nil client here")
	}
}

func TestUnitNewOrchestraMaybeLookupLocationError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)
	clnt, err := sess.NewOrchestraClient(ctx)
	if !strings.HasSuffix(err.Error(), "All IP lookuppers failed") {
		t.Fatal("not the error we expected")
	}
	if clnt != nil {
		t.Fatal("expected nil client here")
	}
}

func TestInitOrchestraClientMaybeRegisterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)
	clnt := orchestra.NewClient(
		sess.HTTPDefaultClient,
		sess.Logger,
		sess.UserAgent(),
		statefile.NewMemory(sess.AssetsDir),
	)
	outclnt, err := sess.initOrchestraClient(
		ctx, clnt, clnt.MaybeLogin,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
	if outclnt != nil {
		t.Fatal("expected a nil client here")
	}
}

func TestInitOrchestraClientMaybeLoginError(t *testing.T) {
	ctx := context.Background()
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)
	clnt := orchestra.NewClient(
		sess.HTTPDefaultClient,
		sess.Logger,
		sess.UserAgent(),
		statefile.NewMemory(sess.AssetsDir),
	)
	expected := errors.New("mocked error")
	outclnt, err := sess.initOrchestraClient(
		ctx, clnt, func(context.Context) error {
			return expected
		},
	)
	if !errors.Is(err, expected) {
		t.Fatal("not the error we expected")
	}
	if outclnt != nil {
		t.Fatal("expected a nil client here")
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
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", URL, nil,
		"../../testdata/",
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
	sess := New(
		log.Log, softwareName, softwareVersion, "../testdata", nil, nil,
		"../../testdata/",
	)
	if err := sess.MaybeLookupLocation(ctx); err == nil {
		t.Fatal("expected an error here")
	}
	if sess.ProbeASNString() != model.DefaultProbeASNString {
		t.Fatal("unexpected ProbeASNString")
	}
	if sess.ProbeASN() != model.DefaultProbeASN {
		t.Fatal("unexpected ProbeASN")
	}
	if sess.ProbeCC() != model.DefaultProbeCC {
		t.Fatal("unexpected ProbeCC")
	}
	if sess.ProbeIP() != model.DefaultProbeIP {
		t.Fatal("unexpected ProbeIP")
	}
	if sess.ProbeNetworkName() != model.DefaultProbeNetworkName {
		t.Fatal("unexpected ProbeNetworkName")
	}
	if sess.ResolverIP() != model.DefaultResolverIP {
		t.Fatal("unexpected ResolverIP")
	}
}
