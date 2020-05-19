package engine

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
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/internal/kvstore"
	"github.com/ooni/probe-engine/internal/orchestra"
	"github.com/ooni/probe-engine/internal/orchestra/statefile"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/probeservices"
)

func TestNewSessionBuilderChecks(t *testing.T) {
	t.Run("with no settings", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{})
	})
	t.Run("with only assets dir", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir: "testdata",
		})
	})
	t.Run("with also logger", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir: "testdata",
			Logger:    log.Log,
		})
	})
	t.Run("with also software name", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir:    "testdata",
			Logger:       log.Log,
			SoftwareName: "ooniprobe-engine",
		})
	})
	t.Run("with also software version", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir:       "testdata",
			Logger:          log.Log,
			SoftwareName:    "ooniprobe-engine",
			SoftwareVersion: "0.0.1",
		})
	})
}

func TestNewSessionBuilderGood(t *testing.T) {
	newSessionForTesting(t)
}

func newSessionMustFail(t *testing.T, config SessionConfig) {
	sess, err := NewSession(config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if sess != nil {
		t.Fatal("expected nil session here")
	}
}

func newSessionForTestingNoLookupsWithProxyURL(t *testing.T, URL *url.URL) *Session {
	tempdir, err := ioutil.TempDir("testdata", "enginetests")
	if err != nil {
		t.Fatal(err)
	}
	sess, err := NewSession(SessionConfig{
		AssetsDir: "testdata",
		AvailableBouncers: []model.Service{{
			Address: "https://ps-test.ooni.io",
			Type:    "https",
		}},
		AvailableCollectors: []model.Service{{
			Address: "https://ps-test.ooni.io",
			Type:    "https",
		}},
		Logger: log.Log,
		PrivacySettings: model.PrivacySettings{
			IncludeASN:     true,
			IncludeCountry: true,
			IncludeIP:      false,
		},
		ProxyURL:        URL,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         tempdir,
	})
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func newSessionForTestingNoLookups(t *testing.T) *Session {
	return newSessionForTestingNoLookupsWithProxyURL(t, nil)
}

func newSessionForTesting(t *testing.T) *Session {
	sess := newSessionForTestingNoLookups(t)
	if err := sess.MaybeLookupLocation(); err != nil {
		t.Fatal(err)
	}
	log.Infof("Platform: %s", sess.Platform())
	log.Infof("ProbeASN: %d", sess.ProbeASN())
	log.Infof("ProbeASNString: %s", sess.ProbeASNString())
	log.Infof("ProbeCC: %s", sess.ProbeCC())
	log.Infof("ProbeIP: %s", sess.ProbeIP())
	log.Infof("ProbeNetworkName: %s", sess.ProbeNetworkName())
	log.Infof("ResolverASN: %d", sess.ResolverASN())
	log.Infof("ResolverASNString: %s", sess.ResolverASNString())
	log.Infof("ResolverIP: %s", sess.ResolverIP())
	log.Infof("ResolverNetworkName: %s", sess.ResolverNetworkName())
	if err := sess.MaybeLookupBackends(); err != nil {
		t.Fatal(err)
	}
	return sess
}

func TestIntegrationNewOrchestraClient(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	clnt, err := sess.NewOrchestraClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if clnt == nil {
		t.Fatal("expected non nil client here")
	}
}

func TestUnitInitOrchestraClientMaybeRegisterError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	clnt := orchestra.NewClient(
		sess.DefaultHTTPClient(),
		sess.Logger(),
		sess.UserAgent(),
		statefile.New(kvstore.NewMemoryKeyValueStore()),
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

func TestUnitInitOrchestraClientMaybeLoginError(t *testing.T) {
	ctx := context.Background()
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	clnt := orchestra.NewClient(
		sess.DefaultHTTPClient(),
		sess.Logger(),
		sess.UserAgent(),
		statefile.New(kvstore.NewMemoryKeyValueStore()),
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
	sess := newSessionForTestingNoLookupsWithProxyURL(t, URL)
	defer sess.Close()
	if sess.ProxyURL() == nil {
		t.Fatal("expected to see explicit proxy here")
	}
	if err := sess.MaybeLookupBackends(); err == nil {
		t.Fatal("expected an error here")
	}
}

func TestIntegrationMaybeStartTunnel(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if err := sess.MaybeStartTunnel(ctx, "psiphon"); err != nil {
		t.Fatal(err)
	}
	if sess.TunnelBootstrapTime() <= 0 {
		t.Fatal("expected positive boostrap time")
	}
	if sess.ProxyURL() == nil {
		t.Fatal("expected non nil proxy URL")
	}
}

func TestIntegrationSessionLocationLookup(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	if err := sess.MaybeLookupLocation(); err != nil {
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
	if sess.ResolverASN() == model.DefaultResolverASN {
		t.Fatal("unexpected ResolverASN")
	}
	if sess.ResolverASNString() == model.DefaultResolverASNString {
		t.Fatal("unexpected ResolverASNString")
	}
	if sess.ResolverIP() == model.DefaultResolverIP {
		t.Fatal("unexpected ResolverIP")
	}
	if sess.ResolverNetworkName() == model.DefaultResolverNetworkName {
		t.Fatal("unexpected ResolverNetworkName")
	}
	if sess.KibiBytesSent() <= 0 {
		t.Fatal("unexpected KibiBytesSent")
	}
	if sess.KibiBytesReceived() <= 0 {
		t.Fatal("unexpected KibiBytesReceived")
	}
}

func TestIntegrationSessionDownloadResources(t *testing.T) {
	tmpdir, err := ioutil.TempDir("testdata", "test-download-resources-idempotent")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	sess.SetAssetsDir(tmpdir)
	err = sess.FetchResourcesIdempotent(ctx)
	if err != nil {
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
}

func TestUnitGetAvailableBouncers(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	all := sess.GetAvailableBouncers()
	diff := cmp.Diff(all, probeservices.Default())
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestUnitMaybeLookupBackendsFailure(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	err = sess.MaybeLookupBackendsContext(ctx)
	if !strings.HasSuffix(err.Error(), "All available bouncers failed") {
		t.Fatal("unexpected error")
	}
}

func TestIntegrationMaybeLookupTestHelpersIdempotent(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	ctx := context.Background()
	if err = sess.MaybeLookupTestHelpersContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err = sess.MaybeLookupTestHelpersContext(ctx); err != nil {
		t.Fatal(err)
	}
	if sess.QueryBouncerCount() != 1 {
		t.Fatal("unexpected number of queries sent to the bouncer")
	}
}

func TestUnitAllBouncersUnsupported(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TempDir:         "testdata",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	sess.AppendAvailableBouncer(model.Service{
		Address: "mascetti",
		Type:    "antani",
	})
	err = sess.MaybeLookupBackends()
	if !strings.HasSuffix(err.Error(), "All available bouncers failed") {
		t.Fatal("unexpected error")
	}
}

func TestIntegrationStartTunnel(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if sess.MaybeStartTunnel(ctx, "") != nil {
		t.Fatal("expected no error here")
	}
	if err := sess.MaybeStartTunnel(ctx, "antani"); err.Error() != "unsupported tunnel" {
		t.Fatal("not the error we expected")
	}
	if err := sess.MaybeStartTunnel(ctx, "psiphon"); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeStartTunnel(ctx, "psiphon"); err != nil {
		t.Fatal(err) // check twice, must be idempotent
	}
}

func TestIntegrationStartTunnelFailure(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := sess.MaybeStartTunnel(ctx, "psiphon")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}
