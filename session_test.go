package engine

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/model"
	"github.com/ooni/probe-engine/netx"
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
	t.Run("with software version and wrong tempdir", func(t *testing.T) {
		newSessionMustFail(t, SessionConfig{
			AssetsDir:       "testdata",
			Logger:          log.Log,
			SoftwareName:    "ooniprobe-engine",
			SoftwareVersion: "0.0.1",
			TempDir:         "./nonexistent",
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

func TestSessionTorArgsTorBinary(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir: "testdata",
		AvailableProbeServices: []model.Service{{
			Address: "https://ams-pg.ooni.org",
			Type:    "https",
		}},
		Logger: log.Log,
		PrivacySettings: model.PrivacySettings{
			IncludeASN:     true,
			IncludeCountry: true,
			IncludeIP:      false,
		},
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
		TorArgs:         []string{"antani1", "antani2", "antani3"},
		TorBinary:       "mascetti",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sess.TorBinary() != "mascetti" {
		t.Fatal("not the TorBinary we expected")
	}
	if len(sess.TorArgs()) != 3 {
		t.Fatal("not the TorArgs length we expected")
	}
	if sess.TorArgs()[0] != "antani1" {
		t.Fatal("not the TorArgs[0] we expected")
	}
	if sess.TorArgs()[1] != "antani2" {
		t.Fatal("not the TorArgs[1] we expected")
	}
	if sess.TorArgs()[2] != "antani3" {
		t.Fatal("not the TorArgs[2] we expected")
	}
}

func newSessionForTestingNoLookupsWithProxyURL(t *testing.T, URL *url.URL) *Session {
	sess, err := NewSession(SessionConfig{
		AssetsDir: "testdata",
		AvailableProbeServices: []model.Service{{
			Address: "https://ams-pg.ooni.org",
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
	})
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

func newSessionForTestingNoLookups(t *testing.T) *Session {
	return newSessionForTestingNoLookupsWithProxyURL(t, nil)
}

func newSessionForTestingNoBackendsLookup(t *testing.T) *Session {
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
	return sess
}

func newSessionForTesting(t *testing.T) *Session {
	sess := newSessionForTestingNoBackendsLookup(t)
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
	clnt, err := probeservices.NewClient(sess, model.Service{
		Address: "https://ams-pg.ooni.org/",
		Type:    "https",
	})
	if err != nil {
		t.Fatal(err)
	}
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
	clnt, err := probeservices.NewClient(sess, model.Service{
		Address: "https://ams-pg.ooni.org/",
		Type:    "https",
	})
	if err != nil {
		t.Fatal(err)
	}
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

func TestMaybeLookupBackendsNewClientError(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	sess.availableProbeServices = []model.Service{{
		Type:    "onion",
		Address: "httpo://jehhrikjjqrlpufu.onion",
	}}
	defer sess.Close()
	err := sess.MaybeLookupBackends()
	if !errors.Is(err, ErrAllProbeServicesFailed) {
		t.Fatal("not the error we expected")
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

func TestIntegrationSessionCloseCancelsTempDir(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	tempDir := sess.TempDir()
	if _, err := os.Stat(tempDir); err != nil {
		t.Fatal(err)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tempDir); !errors.Is(err, syscall.ENOENT) {
		t.Fatal("not the error we expected")
	}
}

func TestIntegrationSessionDownloadResources(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-download-resources-idempotent")
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

func TestUnitGetAvailableProbeServices(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	all := sess.GetAvailableProbeServices()
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
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // so we fail immediately
	err = sess.MaybeLookupBackendsContext(ctx)
	if !errors.Is(err, ErrAllProbeServicesFailed) {
		t.Fatal("unexpected error")
	}
}

func TestIntegrationMaybeLookupTestHelpersIdempotent(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	ctx := context.Background()
	if err = sess.MaybeLookupBackendsContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err = sess.MaybeLookupBackendsContext(ctx); err != nil {
		t.Fatal(err)
	}
	if sess.QueryProbeServicesCount() != 1 {
		t.Fatal("unexpected number of queries sent to the bouncer")
	}
}

func TestUnitAllProbeServicesUnsupported(t *testing.T) {
	sess, err := NewSession(SessionConfig{
		AssetsDir:       "testdata",
		Logger:          log.Log,
		SoftwareName:    "ooniprobe-engine",
		SoftwareVersion: "0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()
	sess.AppendAvailableProbeService(model.Service{
		Address: "mascetti",
		Type:    "antani",
	})
	err = sess.MaybeLookupBackends()
	if !errors.Is(err, ErrAllProbeServicesFailed) {
		t.Fatal("unexpected error")
	}
}

func TestIntegrationStartTunnelGood(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if err := sess.MaybeStartTunnel(ctx, "psiphon"); err != nil {
		t.Fatal(err)
	}
	if err := sess.MaybeStartTunnel(ctx, "psiphon"); err != nil {
		t.Fatal(err) // check twice, must be idempotent
	}
	if sess.ProxyURL() == nil {
		t.Fatal("expected non-nil ProxyURL")
	}
	if sess.TunnelBootstrapTime() <= 0 {
		t.Fatal("expected positive boostrap time")
	}
}

func TestIntegrationStartTunnelNonexistent(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if err := sess.MaybeStartTunnel(ctx, "antani"); err.Error() != "unsupported tunnel" {
		t.Fatal("not the error we expected")
	}
	if sess.ProxyURL() != nil {
		t.Fatal("expected nil ProxyURL")
	}
}

func TestIntegrationStartTunnelEmptyString(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if sess.MaybeStartTunnel(ctx, "") != nil {
		t.Fatal("expected no error here")
	}
	if sess.ProxyURL() != nil {
		t.Fatal("expected nil ProxyURL")
	}
}

func TestIntegrationStartTunnelEmptyStringWithProxy(t *testing.T) {
	proxyURL := &url.URL{Scheme: "socks5", Host: "127.0.0.1:9050"}
	sess := newSessionForTestingNoLookups(t)
	sess.proxyURL = proxyURL
	defer sess.Close()
	ctx := context.Background()
	if sess.MaybeStartTunnel(ctx, "") != nil {
		t.Fatal("expected no error here")
	}
	diff := cmp.Diff(proxyURL, sess.ProxyURL())
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestIntegrationStartTunnelWithAlreadyExistingTunnel(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	if sess.MaybeStartTunnel(ctx, "psiphon") != nil {
		t.Fatal("expected no error here")
	}
	prev := sess.ProxyURL()
	err := sess.MaybeStartTunnel(ctx, "tor")
	if !errors.Is(err, ErrAlreadyUsingProxy) {
		t.Fatal("expected another error here")
	}
	cur := sess.ProxyURL()
	diff := cmp.Diff(prev, cur)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestIntegrationStartTunnelWithAlreadyExistingProxy(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx := context.Background()
	orig := &url.URL{Scheme: "socks5", Host: "[::1]:9050"}
	sess.proxyURL = orig
	err := sess.MaybeStartTunnel(ctx, "psiphon")
	if !errors.Is(err, ErrAlreadyUsingProxy) {
		t.Fatal("expected another error here")
	}
	cur := sess.ProxyURL()
	diff := cmp.Diff(orig, cur)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestIntegrationStartTunnelCanceledContext(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	defer sess.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel
	err := sess.MaybeStartTunnel(ctx, "psiphon")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("not the error we expected")
	}
}

func TestUserAgentNoProxy(t *testing.T) {
	expect := "ooniprobe-engine/0.0.1 ooniprobe-engine/" + Version
	sess := newSessionForTestingNoLookups(t)
	ua := sess.UserAgent()
	diff := cmp.Diff(expect, ua)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestNewOrchestraClientMaybeLookupBackendsFailure(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // fail immediately
	client, err := sess.NewOrchestraClient(ctx)
	if !errors.Is(err, ErrAllProbeServicesFailed) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

type httpTransportThatSleeps struct {
	txp netx.HTTPRoundTripper
	st  time.Duration
}

func (txp httpTransportThatSleeps) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := txp.txp.RoundTrip(req)
	time.Sleep(txp.st)
	return resp, err
}

func (txp httpTransportThatSleeps) CloseIdleConnections() {
	txp.txp.CloseIdleConnections()
}

func TestNewOrchestraClientMaybeLookupLocationFailure(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	sess.httpDefaultTransport = httpTransportThatSleeps{
		txp: sess.httpDefaultTransport,
		st:  5 * time.Second,
	}
	// the transport sleeps for five seconds, so the context should be expired by
	// the time in which we attempt at looking up the clocation
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	client, err := sess.NewOrchestraClient(ctx)
	if err == nil || err.Error() != "All IP lookuppers failed" {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}

func TestNewOrchestraClientProbeServicesNewClientFailure(t *testing.T) {
	sess := newSessionForTestingNoLookups(t)
	sess.selectedProbeServiceHook = func(svc *model.Service) {
		svc.Type = "antani" // should really not be supported for a long time
	}
	client, err := sess.NewOrchestraClient(context.Background())
	if !errors.Is(err, probeservices.ErrUnsupportedEndpoint) {
		t.Fatal("not the error we expected")
	}
	if client != nil {
		t.Fatal("expected nil client here")
	}
}
